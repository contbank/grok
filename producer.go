package grok

import (
	"encoding/json"
	"strings"

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
func (p *MessageBrokerProducer) Publish(topicID string, data interface{}) (string, error) {
	messageId, err := p.PublishWihAttribrutes(topicID, data, nil)
	return messageId, err
}

// PublishWihAttribrutes ...
func (p *MessageBrokerProducer) PublishWihAttribrutes(topicID string, data interface{}, attributes map[string]string) (string, error) {
	body, err := json.Marshal(data)

	if err != nil {
		return "", err
	}

	topic, err := createTopicIfNotExists(p.snsSvc, topicID)

	if err != nil {
		return "", err
	}

	message := string(body)

	output, err := p.snsSvc.Publish(&sns.PublishInput{
		Message:  &message,
		TopicArn: topic,
	})

	if err != nil {
		return "", err
	}

	return *output.MessageId, nil

}

func createTopicIfNotExists(snsSvc *sns.SNS, id string) (*string, error) {
	var topicArn *string

	allTopics, _ := snsSvc.ListTopics(&sns.ListTopicsInput{})

	for _, t := range allTopics.Topics {
		splitTopic := strings.Split(*t.TopicArn, ":")
		strTopic := splitTopic[len(splitTopic)-1]
		if strings.Compare(strTopic, id) == 0 {
			topicArn = t.TopicArn
			break
		}
	}

	if topicArn != nil {
		return topicArn, nil
	}

	topic, err := snsSvc.CreateTopic(&sns.CreateTopicInput{
		Name: aws.String(id),
	})

	if err != nil {
		return nil, err
	}

	return topic.TopicArn, nil
}
