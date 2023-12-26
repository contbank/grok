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
	sqsSvc       *sqs.SQS
	snsSvc       *sns.SNS
	handler      func(interface{}) error
	subscriberID string
	topicID      string
	topicIDs     []string
	handleType   reflect.Type
	maxRetries   int
	fifo         *bool
}

// MessageBrokerSubscriberOption ...
type MessageBrokerSubscriberOption func(*MessageBrokerSubscriber)

// NewMessageBrokerSubscriber ...
func NewMessageBrokerSubscriber(opts ...MessageBrokerSubscriberOption) *MessageBrokerSubscriber {
	subscriber := new(MessageBrokerSubscriber)
	subscriber.maxRetries = 5
	subscriber.topicIDs = make([]string, 0)

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

// método WithTopicID para aceitar múltiplos IDs de tópicos
func WithTopicID(ids ...string) MessageBrokerSubscriberOption {
	return func(s *MessageBrokerSubscriber) {
		s.topicIDs = append(s.topicIDs, ids...)
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

// WithFIFO - default false
func WithFIFO(fifo *bool) MessageBrokerSubscriberOption {
	return func(s *MessageBrokerSubscriber) {
		s.fifo = fifo
	}
}

// WithFIFOAttributes ...
func WithFIFOAttributes(messageGroupID *string, messageDeduplicationID *string) map[string]string {
	return map[string]string{
		"Fifo":                   strconv.FormatBool(true),
		"MessageGroupID":         *messageGroupID,
		"MessageDeduplicationID": *messageDeduplicationID,
	}
}

// Run ...
func (s *MessageBrokerSubscriber) Run() error {

	if len(s.topicIDs) == 0 {
		err := NewError(404, "SUBSCRIBER_ERROR", "error in topic subscriber")
		return err
	}
	var queueURL *string
	var err error
	for _, topicID := range s.topicIDs {
		queueURL, err = createSubscriptionIfNotExists(s.sqsSvc, s.snsSvc, s.subscriberID, topicID, s.maxRetries, s.fifo)

		if err != nil {
			logrus.WithError(err).
				Errorf("error starting %s", s.subscriberID)
			continue
		}
		logrus.Infof("starting consumer %s with topic %s", s.subscriberID, topicID)
	}

	if err := s.checkMessages(s.sqsSvc, queueURL); err != nil {
		return err
	}
	return nil
}

func createSubscriptionIfNotExists(sqsSvc *sqs.SQS, snsSvc *sns.SNS, subscriberID, topicID string, maxRetries int, fifo *bool) (*string, error) {
	listQueueResults, err := sqsSvc.ListQueues(&sqs.ListQueuesInput{
		QueueNamePrefix: aws.String(subscriberID),
	})

	if err != nil {
		logrus.WithError(err).
			Errorf("error list queues %s", subscriberID)
		return nil, err
	}

	var queueURL *string
	var queueDlqURL *string

	for _, t := range listQueueResults.QueueUrls {
		parts := strings.Split(*t, "/")
		if strings.Compare(parts[4], subscriberID) == 0 {
			queueURL = t
			break
		}
	}

	if queueURL == nil {
		sqsName := subscriberID
		sqsAttributes := map[string]*string{
			sqs.QueueAttributeNameReceiveMessageWaitTimeSeconds: aws.String("20"),
		}

		if fifo != nil && *fifo {
			stringFifo := strconv.FormatBool(*fifo)
			sqsAttributes[sqs.QueueAttributeNameFifoQueue] = &stringFifo
			sqsAttributes[sqs.QueueAttributeNameContentBasedDeduplication] = &stringFifo
			sqsName = fmt.Sprintf("%s.fifo", sqsName)
		}

		resp, err := sqsSvc.CreateQueue(&sqs.CreateQueueInput{
			QueueName:  aws.String(sqsName),
			Attributes: sqsAttributes,
		})

		if err != nil {
			logrus.WithError(err).
				Errorf("error creating queue %s", subscriberID)
			return nil, err
		}

		queueURL = resp.QueueUrl

	}

	for _, t := range listQueueResults.QueueUrls {
		parts := strings.Split(*t, "/")
		if strings.Compare(parts[4], fmt.Sprintf("%s_dlq", subscriberID)) == 0 {
			queueDlqURL = t
			break
		}
	}

	if queueDlqURL == nil {

		dlqName := fmt.Sprintf("%s_dlq", subscriberID)
		sqsDlqAttributes := map[string]*string{
			sqs.QueueAttributeNameReceiveMessageWaitTimeSeconds: aws.String("20"),
		}

		if fifo != nil && *fifo {
			stringFifo := strconv.FormatBool(*fifo)
			sqsDlqAttributes[sqs.QueueAttributeNameFifoQueue] = &stringFifo
			sqsDlqAttributes[sqs.QueueAttributeNameContentBasedDeduplication] = &stringFifo
			dlqName = fmt.Sprintf("%s.fifo", dlqName)
		}

		respdlq, err := sqsSvc.CreateQueue(&sqs.CreateQueueInput{
			QueueName:  &dlqName,
			Attributes: sqsDlqAttributes,
		})

		if err != nil {
			logrus.WithError(err).
				Errorf("error creating queue dlq %s", subscriberID)
			return nil, err
		}

		queueDlqURL = respdlq.QueueUrl
	}

	queueARN := convertQueueURLToARN(*queueURL)
	queueDlqARN := convertQueueURLToARN(*queueDlqURL)

	var attributes = make(map[string]string)
	if fifo != nil && *fifo {
		attributes["Fifo"] = strconv.FormatBool(*fifo)
	}

	topicArn, err := createTopicIfNotExists(snsSvc, topicID, attributes)

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
		"deadLetterTargetArn": queueDlqARN,
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
			sqs.QueueAttributeNamePolicy:                        aws.String(policyContent),
			sqs.QueueAttributeNameRedrivePolicy:                 aws.String(string(redrivePolicyContent)),
			sqs.QueueAttributeNameReceiveMessageWaitTimeSeconds: aws.String("20"),
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
