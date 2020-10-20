package grok

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/sirupsen/logrus"
)

// PubSubSubscriber ...
type PubSubSubscriber struct {
	sqsSvc                 *sqs.SQS
	snsSvc                 *sns.SNS
	handler                func(interface{}) error
	subscriberID           string
	topicID                string
	handleType             reflect.Type
	maxRetries             int
	producer               *PubSubProducer
	maxRetriesAttribute    string
	maxOutstandingMessages int
	ackDeadline            time.Duration
}

// PubSubSubscriberOption ...
type PubSubSubscriberOption func(*PubSubSubscriber)

// NewPubSubSubscriber ...
func NewPubSubSubscriber(opts ...PubSubSubscriberOption) *PubSubSubscriber {
	subscriber := new(PubSubSubscriber)
	subscriber.maxRetries = 5
	subscriber.maxOutstandingMessages = pubsub.DefaultReceiveSettings.MaxOutstandingMessages
	subscriber.ackDeadline = 10 * time.Second

	for _, opt := range opts {
		opt(subscriber)
	}

	return subscriber
}

// WithSessionSQS ...
func WithSessionSQS(sessionSQS *session.Session) PubSubSubscriberOption {
	return func(s *PubSubSubscriber) {
		s.sqsSvc = sqs.New(sessionSQS)
	}
}

// WithSessionSNS ...
func WithSessionSNS(sessionSNS *session.Session) PubSubSubscriberOption {
	return func(s *PubSubSubscriber) {
		s.snsSvc = sns.New(sessionSNS)
	}
}

// WithHandler ...
func WithHandler(h func(interface{}) error) PubSubSubscriberOption {
	return func(s *PubSubSubscriber) {
		s.handler = h
	}
}

// WithPubSubSubscriberID ...
func WithPubSubSubscriberID(id string) PubSubSubscriberOption {
	return func(s *PubSubSubscriber) {
		s.subscriberID = id
	}
}

// WithTopicID ...
func WithTopicID(t string) PubSubSubscriberOption {
	return func(s *PubSubSubscriber) {
		s.topicID = t
	}
}

// WithType ...
func WithType(t reflect.Type) PubSubSubscriberOption {
	return func(s *PubSubSubscriber) {
		s.handleType = t
	}
}

// WithMaxRetries - default 5
func WithMaxRetries(maxRetries int) PubSubSubscriberOption {
	return func(s *PubSubSubscriber) {
		s.maxRetries = maxRetries
	}
}

//WithMaxOutstandingMessages ...
func WithMaxOutstandingMessages(maxOutstandingMessages int) PubSubSubscriberOption {
	return func(s *PubSubSubscriber) {
		s.maxOutstandingMessages = maxOutstandingMessages
	}
}

//WithAckDeadline ...
func WithAckDeadline(t time.Duration) PubSubSubscriberOption {
	return func(s *PubSubSubscriber) {
		s.ackDeadline = t
	}
}

// Run ...
func (s *PubSubSubscriber) Run() {
	queueURL, err := createSubscriptionIfNotExists(s.sqsSvc, s.snsSvc, s.subscriberID, s.topicID, s.ackDeadline)

	if err != nil {
		logrus.WithError(err).
			Errorf("error starting %s", s.subscriberID)
	}

	logrus.Infof("starting consumer %s with topic %s", s.subscriberID, s.topicID)
	go s.checkMessages(s.sqsSvc, queueURL)
}

func createSubscriptionIfNotExists(sqsSvc *sqs.SQS, snsSvc *sns.SNS, subscriberID, topicID string, ackDeadline time.Duration) (*string, error) {
	listQueueResults, _ := sqsSvc.ListQueues(&sqs.ListQueuesInput{
		QueueNamePrefix: aws.String(subscriberID),
	})

	var queueURL *string

	for _, t := range listQueueResults.QueueUrls {
		if strings.Contains(*t, subscriberID) {
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

	topicArn, err := createTopicIfNotExists(snsSvc, topicID)

	if err != nil {
		logrus.WithError(err).
			Errorf("error creating topic %s", topicID)
		return nil, err
	}

	queueARN := convertQueueURLToARN(*queueURL)

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

	attr := make(map[string]*string, 1)
	attr["Policy"] = &policyContent

	setQueueAttrInput := sqs.SetQueueAttributesInput{
		QueueUrl:   queueURL,
		Attributes: attr,
	}

	_, err = sqsSvc.SetQueueAttributes(&setQueueAttrInput)

	if err != nil {
		logrus.WithError(err).
			Errorf("error set attributes policy queue %s", subscriberID)
		return nil, err
	}

	return queueURL, nil
}

func (s *PubSubSubscriber) retry(message *pubsub.Message, body interface{}) error {
	retries := s.getRetries(message)
	retries++

	message.Attributes[s.maxRetriesAttribute] = strconv.Itoa(retries)

	return s.producer.PublishWihAttribrutes(s.topicID, body, message.Attributes)
}

func (s *PubSubSubscriber) dlq(message *pubsub.Message, e error) error {
	dlq := fmt.Sprintf("%s_dlq", s.topicID)

	logrus.Infof("sending message %s to %s", message.ID, dlq)

	_, err := createTopicIfNotExists(s.snsSvc, dlq)

	if err != nil {
		return err
	}

	attributes := make(map[string]string)
	attributes["error"] = e.Error()

	return s.producer.PublishWihAttribrutes(dlq, message.Data, attributes)
}

func (s *PubSubSubscriber) getRetries(message *pubsub.Message) int {
	if message.Attributes == nil {
		message.Attributes = make(map[string]string)
	}

	retries := 0
	attribute, ok := message.Attributes[s.maxRetriesAttribute]

	if ok {
		retries, _ = strconv.Atoi(attribute)
	}

	return retries
}

func (s *PubSubSubscriber) checkMessages(sqsSvc *sqs.SQS, queueURL *string) {
	for {
		retrieveMessageRequest := sqs.ReceiveMessageInput{
			QueueUrl: queueURL,
		}

		retrieveMessageResponse, _ := sqsSvc.ReceiveMessage(&retrieveMessageRequest)

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
	queueARN := strings.Replace(strings.Replace(strings.Replace(inputURL, "https://sqs.", "arn:aws:sqs:", -1), ".amazonaws.com/", ":", -1), "/", ":", -1)
	queueARN = strings.Replace(strings.Replace(strings.Replace(queueARN, "https://sqs.", "arn:aws:sqs:", -1), ".localhost/", ":", -1), "/", ":", -1)

	return queueARN
}
