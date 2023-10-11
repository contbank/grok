package grok

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

// MessageBrokerProducer ...
type MessageBrokerProducer struct {
	snsSvc *sns.SNS
}

// NewMessageBrokerProducer ...
func NewMessageBrokerProducer(s *session.Session) *MessageBrokerProducer {
	snsSvc := sns.New(s)
	return &MessageBrokerProducer{snsSvc: snsSvc}
}

// Publish ...
func (p *MessageBrokerProducer) Publish(topicID string, data interface{}, attributes map[string]string) (string, error) {
	messageId, err := p.PublishWihAttributes(topicID, data, attributes)
	return messageId, err
}

// PublishMany ...
func (p *MessageBrokerProducer) PublishMany(topics []string, data interface{}) (map[string]string, map[string]error) {
	publishErrors := make(map[string]error, len(topics))
	publishOk := make(map[string]string, len(topics))

	for _, topicName := range topics {
		messageId, err := p.Publish(topicName, data, nil)
		if err != nil {
			publishErrors[topicName] = err
			logrus.WithError(err).
				Errorf("failed to send message to %s - send card event", topicName)
			break
		}
		logrus.Infof("sent card event to %s. message id %s", topicName, messageId)
		publishOk[topicName] = messageId
	}

	return publishOk, publishErrors
}

// PublishWihAttribrutes ...
func (p *MessageBrokerProducer) PublishWihAttributes(topicID string, data interface{}, attributes map[string]string) (string, error) {
	body, err := json.Marshal(data)

	if err != nil {
		return "", err
	}

	topic, err := createTopicIfNotExists(p.snsSvc, topicID, attributes)

	if err != nil {
		return "", err
	}

	message := string(body)

	snsPublishInput := &sns.PublishInput{
		Message:  &message,
		TopicArn: topic,
	}

	if len(attributes) != 0 {
		if value, ok := attributes["Fifo"]; ok {
			fifo, err := strconv.ParseBool(value)
			if err != nil {
				return "", err
			}
			if fifo {
				if value, ok := attributes["MessageGroupID"]; ok {
					snsPublishInput.MessageAttributes = make(map[string]*sns.MessageAttributeValue)
					snsPublishInput.MessageAttributes["MessageGroupId"] = &sns.MessageAttributeValue{
						DataType:    aws.String("String"),
						StringValue: aws.String(value),
					}
					snsPublishInput.MessageGroupId = &value
				}
				if value, ok := attributes["MessageDeduplicationID"]; ok {
					snsPublishInput.MessageAttributes = make(map[string]*sns.MessageAttributeValue)
					snsPublishInput.MessageAttributes["MessageDeduplicationId"] = &sns.MessageAttributeValue{
						DataType:    aws.String("String"),
						StringValue: aws.String(value),
					}
					snsPublishInput.MessageDeduplicationId = &value
				}
			}
		}
	}

	output, err := p.snsSvc.Publish(snsPublishInput)
	if err != nil {
		return "", err
	}

	return *output.MessageId, nil

}

func createTopicIfNotExists(snsSvc *sns.SNS, id string, attributes map[string]string) (*string, error) {
	var topicArn *string
	snsName := id
	snsAttributes := map[string]*string{}
	if value, ok := attributes["Fifo"]; ok {
		fifo, err := strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}

		snsAttributes["FifoTopic"] = &value
		snsAttributes["ContentBasedDeduplication"] = &value
		if fifo {
			snsName = fmt.Sprintf("%s.fifo", snsName)
		}
	}

	allTopics, err := snsSvc.ListTopics(&sns.ListTopicsInput{})
	if err != nil {
		return nil, err
	}

	for _, t := range allTopics.Topics {
		splitTopic := strings.Split(*t.TopicArn, ":")
		strTopic := splitTopic[len(splitTopic)-1]
		if strings.Compare(strTopic, snsName) == 0 {
			topicArn = t.TopicArn
			break
		}
	}

	if topicArn != nil {
		return topicArn, nil
	}

	topic, err := snsSvc.CreateTopic(&sns.CreateTopicInput{
		Name:       aws.String(snsName),
		Attributes: snsAttributes,
	})

	if err != nil {
		return nil, err
	}

	return topic.TopicArn, nil
}
