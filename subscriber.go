package grok

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/sirupsen/logrus"
)

// MessageBrokerSubscriber ...
type MessageBrokerSubscriber struct {
	sqsSvc              *sqs.SQS
	snsSvc              *sns.SNS
	handler             func(interface{}) error
	subscriberID        string
	topicID             string
	handleType          reflect.Type
	maxRetries          int
	producer            *MessageBrokerProducer
	maxRetriesAttribute string
}

// MessageBrokerSubscriberOption ...
type MessageBrokerSubscriberOption func(*MessageBrokerSubscriber)

// NewMessageBrokerSubscriber ...
func NewMessageBrokerSubscriber(opts ...MessageBrokerSubscriberOption) *MessageBrokerSubscriber {
	subscriber := new(MessageBrokerSubscriber)
	subscriber.maxRetries = 5

	for _, opt := range opts {
		opt(subscriber)
	}

	return subscriber
}

// WithSessionSQS ...
func WithSessionSQS(sessionSQS *session.Session) MessageBrokerSubscriberOption {
	return func(s *MessageBrokerSubscriber) {
		s.sqsSvc = sqs.New(sessionSQS)
	}
}

// WithSessionSNS ...
func WithSessionSNS(sessionSNS *session.Session) MessageBrokerSubscriberOption {
	return func(s *MessageBrokerSubscriber) {
		s.snsSvc = sns.New(sessionSNS)
	}
}

// WithHandler ...
func WithHandler(h func(interface{}) error) MessageBrokerSubscriberOption {
	return func(s *MessageBrokerSubscriber) {
		s.handler = h
	}
}

// WithSubscriberID ...
func WithSubscriberID(id string) MessageBrokerSubscriberOption {
	return func(s *MessageBrokerSubscriber) {
		s.subscriberID = id
	}
}

// WithTopicID ...
func WithTopicID(t string) MessageBrokerSubscriberOption {
	return func(s *MessageBrokerSubscriber) {
		s.topicID = t
	}
}

// WithType ...
func WithType(t reflect.Type) MessageBrokerSubscriberOption {
	return func(s *MessageBrokerSubscriber) {
		s.handleType = t
	}
}

// WithMaxRetries - default 5
func WithMaxRetries(maxRetries int) MessageBrokerSubscriberOption {
	return func(s *MessageBrokerSubscriber) {
		s.maxRetries = maxRetries
	}
}

// Run ...
func (s *MessageBrokerSubscriber) Run() error {
	queueURL, err := createSubscriptionIfNotExists(s.sqsSvc, s.snsSvc, s.subscriberID, s.topicID, s.maxRetries)

	if err != nil {
		logrus.WithError(err).
			Errorf("error starting %s", s.subscriberID)
	}

	logrus.Infof("starting consumer %s with topic %s", s.subscriberID, s.topicID)
	return s.checkMessages(s.sqsSvc, queueURL)
}

func createSubscriptionIfNotExists(sqsSvc *sqs.SQS, snsSvc *sns.SNS, subscriberID, topicID string, maxRetries int) (*string, error) {
	listQueueResults, err := sqsSvc.ListQueues(&sqs.ListQueuesInput{
		QueueNamePrefix: aws.String(subscriberID),
	})

	if err != nil {
		logrus.WithError(err).
			Errorf("error list queues %s", subscriberID)
		return nil, err
	}

	var queueURL *string

	for _, t := range listQueueResults.QueueUrls {
		parts := strings.Split(*t, "/")
		if strings.Compare(parts[4], subscriberID) == 0 {
			queueURL = t
			break
		}
	}

	if queueURL != nil {
		return queueURL, nil
	}

	resp, err := sqsSvc.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String(subscriberID),
	})

	if err != nil {
		logrus.WithError(err).
			Errorf("error creating queue %s", subscriberID)
		return nil, err
	}

	queueURL = resp.QueueUrl
	queueARN := convertQueueURLToARN(*queueURL)

	respdlq, err := sqsSvc.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String(fmt.Sprintf("%s_dlq", subscriberID)),
	})

	if err != nil {
		logrus.WithError(err).
			Errorf("error creating queue dlq %s", subscriberID)
		return nil, err
	}

	dlqURL := respdlq.QueueUrl
	dlqARN := convertQueueURLToARN(*dlqURL)

	topicArn, err := createTopicIfNotExists(snsSvc, topicID)

	if err != nil {
		logrus.WithError(err).
			Errorf("error creating topic %s", topicID)
		return nil, err
	}

	_, err = snsSvc.Subscribe(&sns.SubscribeInput{
		TopicArn: topicArn,
		Protocol: aws.String("sqs"),
		Endpoint: &queueARN,
	})

	if err != nil {
		logrus.WithError(err).
			Errorf("error subscribe topic %s", topicID)
		return nil, err
	}

	policyContent := "{\"Version\": \"2012-10-17\",  \"Id\": \"" + queueARN + "/SQSDefaultPolicy\",  \"Statement\": [    {     \"Sid\": \"Sid1580665629194\",      \"Effect\": \"Allow\",      \"Principal\": {        \"AWS\": \"*\"      },      \"Action\": \"SQS:SendMessage\",      \"Resource\": \"" + queueARN + "\",      \"Condition\": {        \"ArnEquals\": {         \"aws:SourceArn\": \"" + *topicArn + "\"        }      }    }  ]}"

	policy := map[string]string{
		"deadLetterTargetArn": dlqARN,
		"maxReceiveCount":     strconv.Itoa(maxRetries),
	}

	redrivePolicyContent, err := json.Marshal(policy)

	if err != nil {
		logrus.WithError(err).
			Errorf("error marshal redrive policy %s", subscriberID)
		return nil, err
	}

	setQueueAttrInput := sqs.SetQueueAttributesInput{
		QueueUrl: queueURL,
		Attributes: map[string]*string{
			sqs.QueueAttributeNamePolicy:        aws.String(policyContent),
			sqs.QueueAttributeNameRedrivePolicy: aws.String(string(redrivePolicyContent)),
		},
	}

	_, err = sqsSvc.SetQueueAttributes(&setQueueAttrInput)

	if err != nil {
		logrus.WithError(err).
			Errorf("error set attributes policy queue %s", subscriberID)
		return nil, err
	}

	return queueURL, nil
}

func (s *MessageBrokerSubscriber) checkMessages(sqsSvc *sqs.SQS, queueURL *string) error {
	for {
		retrieveMessageRequest := sqs.ReceiveMessageInput{
			QueueUrl: queueURL,
		}

		retrieveMessageResponse, err := sqsSvc.ReceiveMessage(&retrieveMessageRequest)

		if err != nil {
			logrus.WithError(err).
				Errorf("error receive message")
			return err
		}

		if len(retrieveMessageResponse.Messages) > 0 {

			processedReceiptHandles := make([]*sqs.DeleteMessageBatchRequestEntry, len(retrieveMessageResponse.Messages))

			for i, mess := range retrieveMessageResponse.Messages {
				byt := []byte(*mess.Body)
				body := reflect.New(reflect.TypeOf(map[string]interface{}{})).Interface()
				json.Unmarshal(byt, body)

				value := body.(*map[string]interface{})
				result := (*value)["Message"]
				messageStr := fmt.Sprintf("%v", result)

				bytMessage := []byte(messageStr)
				bodyMessage := reflect.New(s.handleType).Interface()
				err := json.Unmarshal(bytMessage, bodyMessage)

				if err != nil {
					logrus.WithError(err).WithField("content", mess.String()).
						Errorf("cannot unmarshal message %s - sending to dlq", *mess.MessageId)

					processedReceiptHandles[i] = &sqs.DeleteMessageBatchRequestEntry{
						Id:            mess.MessageId,
						ReceiptHandle: mess.ReceiptHandle,
					}
				} else {
					err = s.handler(bodyMessage)

					if err == nil {
						processedReceiptHandles[i] = &sqs.DeleteMessageBatchRequestEntry{
							Id:            mess.MessageId,
							ReceiptHandle: mess.ReceiptHandle,
						}
					}
				}

			}

			deleteMessageRequest := sqs.DeleteMessageBatchInput{
				QueueUrl: queueURL,
				Entries:  processedReceiptHandles,
			}

			sqsSvc.DeleteMessageBatch(&deleteMessageRequest)

		}
	}

}

func convertQueueURLToARN(inputURL string) string {
	var queueARN string
	if strings.Contains(inputURL, "localhost") {
		queueARN = strings.Replace(strings.Replace(strings.Replace(inputURL, "http://", "arn:aws:sqs:", -1), "localhost:4566", "us-west-2", -1), "/", ":", -1)
	} else {
		queueARN = strings.Replace(strings.Replace(strings.Replace(inputURL, "https://sqs.", "arn:aws:sqs:", -1), ".amazonaws.com/", ":", -1), "/", ":", -1)
	}

	return queueARN
}
