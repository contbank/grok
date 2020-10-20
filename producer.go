package grok

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

// PubSubProducer ...
type PubSubProducer struct {
	snsSvc *sns.SNS
}

// NewPubSubProducer ...
func NewPubSubProducer(s *session.Session) *PubSubProducer {
	snsSvc := sns.New(s)
	return &PubSubProducer{snsSvc: snsSvc}
}

// Publish ...
func (p *PubSubProducer) Publish(topicID string, data interface{}) error {
	return p.PublishWihAttribrutes(topicID, data, nil)
}

// PublishWihAttribrutes ...
func (p *PubSubProducer) PublishWihAttribrutes(topicID string, data interface{}, attributes map[string]string) error {
	body, err := json.Marshal(data)

	if err != nil {
		return err
	}

	topic, err := createTopicIfNotExists(p.snsSvc, topicID)

	if err != nil {
		return err
	}

	message := string(body)

	_, err = p.snsSvc.Publish(&sns.PublishInput{
		Message:  &message,
		TopicArn: topic,
	})

	return err
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
