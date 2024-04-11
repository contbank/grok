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
	topicIDs     []string
	handleType   reflect.Type
	maxRetries   int
	fifo         bool
	dlq          bool
}

// MessageBrokerSubscriberOption ...
type MessageBrokerSubscriberOption func(*MessageBrokerSubscriber)

// NewMessageBrokerSubscriber ...
func NewMessageBrokerSubscriber(opts ...MessageBrokerSubscriberOption) *MessageBrokerSubscriber {
	subscriber := new(MessageBrokerSubscriber)
	subscriber.maxRetries = 5
	subscriber.topicIDs = make([]string, 0)
	subscriber.fifo = false
	subscriber.dlq = true

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

// WithTopicID ... método WithTopicID para aceitar múltiplos IDs de tópicos
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
func WithFIFO(fifo bool) MessageBrokerSubscriberOption {
	return func(s *MessageBrokerSubscriber) {
		s.fifo = fifo
	}
}

// WithDLQ - default true
func WithDLQ(dlq bool) MessageBrokerSubscriberOption {
	return func(s *MessageBrokerSubscriber) {
		s.dlq = dlq
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

	var queueURL *string

	if s.dlq {
		if len(s.topicIDs) == 0 {
			err := NewError(404, "SUBSCRIBER_ERROR", "error in topic subscriber")
			return err
		}
		var err error
		for _, topicID := range s.topicIDs {
			queueURL, err = s.createSubscriptionIfNotExists(s.sqsSvc, s.snsSvc, s.subscriberID, topicID)

			if err != nil {
				logrus.WithError(err).
					Errorf("error starting %s", s.subscriberID)
				continue
			}
			logrus.Infof("starting consumer %s with topic %s", s.subscriberID, topicID)
		}
	} else {
		dlqQueueURL, err := s.listQueuesBySubscriberID(s.sqsSvc, s.subscriberID)
		if err != nil {
			logrus.WithError(err).
				Errorf("error starting %s", s.subscriberID)
			return err
		}
		queueURL = dlqQueueURL

		logrus.Infof("starting dlq consumer with queue %s", s.subscriberID)
	}

	if err := s.checkMessages(s.sqsSvc, queueURL); err != nil {
		return err
	}
	return nil
}

func (s *MessageBrokerSubscriber) listQueuesBySubscriberID(sqsSvc *sqs.SQS, subscriberID string) (*string, error) {
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
	return queueURL, nil
}

func (s *MessageBrokerSubscriber) createSubscriptionIfNotExists(sqsSvc *sqs.SQS, snsSvc *sns.SNS, subscriberID, topicID string) (*string, error) {
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

	if s.dlq {

		if queueURL == nil {
			sqsName := subscriberID
			sqsAttributes := map[string]*string{
				sqs.QueueAttributeNameReceiveMessageWaitTimeSeconds:   aws.String("20"),
				sqs.MessageSystemAttributeNameApproximateReceiveCount: aws.String("true"),
			}

			if s.fifo {
				stringFifo := strconv.FormatBool(s.fifo)
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

		if queueDlqURL == nil && s.dlq {

			dlqName := fmt.Sprintf("%s_dlq", subscriberID)
			sqsDlqAttributes := map[string]*string{
				sqs.QueueAttributeNameReceiveMessageWaitTimeSeconds: aws.String("20"),
			}

			if s.fifo {
				stringFifo := strconv.FormatBool(s.fifo)
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

		queueARN := s.convertQueueURLToARN(*queueURL)
		queueDlqARN := s.convertQueueURLToARN(*queueDlqURL)

		var attributes = make(map[string]string)
		if s.fifo {
			attributes["Fifo"] = strconv.FormatBool(s.fifo)
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

		policyContentMap := map[string]interface{}{
			"Version": "2012-10-17",
			"Id":      queueARN + "/SQSDefaultPolicy",
			"Statement": []map[string]interface{}{
				{
					"Sid":       "Sid1580665629194",
					"Effect":    "Allow",
					"Principal": map[string]string{"AWS": "*"},
					"Action":    "SQS:SendMessage",
					"Resource":  queueARN,
					"Condition": map[string]map[string]string{
						"ArnEquals": {"aws:SourceArn": *topicArn},
					},
				},
			},
		}

		policyContent, err := json.Marshal(policyContentMap)
		if err != nil {
			logrus.WithError(err).Errorf("error marshal policy %s", subscriberID)
			return nil, err
		}
		policyContentString := string(policyContent)

		policy := map[string]string{
			"deadLetterTargetArn": queueDlqARN,
			"maxReceiveCount":     strconv.Itoa(s.maxRetries),
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
				sqs.QueueAttributeNamePolicy:                        aws.String(policyContentString),
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

/*
func (s *MessageBrokerSubscriber) parseMessagePayload(message *sqs.Message) (map[string]interface{}, error) {
	var payload map[string]interface{}

	err := json.Unmarshal([]byte(*message.Body), &payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func (s *MessageBrokerSubscriber) moveToDeadLetterQueue(sqsSvc *sqs.SQS, queueURL *string,
	message *sqs.Message) (*sqs.DeleteMessageBatchOutput, error) {
	deleteMessageRequest := &sqs.DeleteMessageBatchInput{
		QueueUrl: queueURL,
		Entries: []*sqs.DeleteMessageBatchRequestEntry{
			{
				Id:            message.MessageId,
				ReceiptHandle: message.ReceiptHandle,
			},
		},
	}

	return sqsSvc.DeleteMessageBatch(deleteMessageRequest)
}

func (s *MessageBrokerSubscriber) deleteMessage(sqsSvc *sqs.SQS, queueURL *string, receiptHandle *string) error {
	deleteMessageRequest := &sqs.DeleteMessageInput{
		QueueUrl:      queueURL,
		ReceiptHandle: receiptHandle,
	}

	_, err := sqsSvc.DeleteMessage(deleteMessageRequest)
	return err
}

func (s *MessageBrokerSubscriber) getApproximateReceiveCount(message *sqs.Message) (int, error) {
	receiveCountStr, ok := message.Attributes["ApproximateReceiveCount"]
	if !ok {
		err := errors.New("error getting ApproximateReceiveCount attribute")
		logrus.WithError(err).
			Errorf("error getting ApproximateReceiveCount attribute.")
		return 0, err
	}

	receiveCount, err := strconv.Atoi(*receiveCountStr)
	if err != nil {
		return 0, err
	}

	return receiveCount, nil
}*/

func (s *MessageBrokerSubscriber) convertQueueURLToARN(inputURL string) string {
	const sqsPrefix = "http://"
	const sqsARNPrefix = "arn:aws:sqs:"
	const localStackEndpoint = "localhost:4566"

	if strings.Contains(inputURL, localStackEndpoint) {
		inputURL = strings.Replace(inputURL, sqsPrefix, sqsARNPrefix, 1)
		inputURL = strings.Replace(inputURL, localStackEndpoint, "us-west-2", 1)
	} else {
		inputURL = strings.Replace(inputURL, "https://sqs.", sqsARNPrefix, 1)
		inputURL = strings.Replace(inputURL, ".amazonaws.com/", ":", 1)
	}

	return strings.Replace(inputURL, "/", ":", -1)
}
