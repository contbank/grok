package grok_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/contbank/grok"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MessageBrokerSubscriberTestSuite struct {
	suite.Suite
	assert     *assert.Assertions
	settings   *grok.Settings
	sessionSQS *session.Session
	sessionSNS *session.Session
	producer   *grok.MessageBrokerProducer
}

func TestMessageBrokerSubscriberTestSuite(t *testing.T) {
	suite.Run(t, new(MessageBrokerSubscriberTestSuite))
}

func (s *MessageBrokerSubscriberTestSuite) SetupTest() {
	s.assert = assert.New(s.T())
	s.settings = &grok.Settings{}
	grok.FromYAML("tests/config.yaml", s.settings)
	s.sessionSQS = grok.CreateSession(s.settings.AWS.SQS)
	s.sessionSNS = grok.CreateSession(s.settings.AWS.SNS)
	s.producer = grok.NewMessageBrokerProducer(s.sessionSNS)

}

func (s *MessageBrokerSubscriberTestSuite) TestSubscribe() {
	received := make(chan bool, 1)

	subscriberID := "subs"
	topicID := "topic-teste"

	message := map[string]interface{}{"ping": "pong"}

	go func() {
		messageBroker := grok.NewMessageBrokerSubscriber(
			grok.WithSessionSQS(s.sessionSQS),
			grok.WithSessionSNS(s.sessionSNS),
			grok.WithTopicID(topicID),
			grok.WithSubscriberID(subscriberID),
			grok.WithType(reflect.TypeOf(message)),
			grok.WithHandler(func(data interface{}) error {
				defer func() { received <- true }()

				value, ok := data.(*map[string]interface{})
				s.assert.True(ok)
				s.assert.Equal("pong", (*value)["ping"])

				return nil
			}),
		)

		err := messageBroker.Run()

		s.assert.NoError(err)
	}()

	time.Sleep(time.Second * 3)

	messageId, err := s.producer.Publish(topicID, message, nil)
	if err != nil {
		received <- true
	}

	s.assert.NoError(err)
	s.assert.NotNil(messageId)

	<-received
}

func (s *MessageBrokerSubscriberTestSuite) TestSubscribes() {
	received := make(chan bool, 2) // Alterado para receber dois sinais

	subscriberID := "subsc"

	message := map[string]interface{}{"ping": "pong"}

	topics := []string{"topic-teste", "topic-teste-dois"} // Tópicos a serem inscritos

	messageBroker := grok.NewMessageBrokerSubscriber(
		grok.WithSessionSQS(s.sessionSQS),
		grok.WithSessionSNS(s.sessionSNS),
		grok.WithSubscriberID(subscriberID),
		grok.WithType(reflect.TypeOf(message)),
		grok.WithTopicID(topics...), // Usando os tópicos especificados
		grok.WithHandler(func(data interface{}) error {
			defer func() { received <- true }()

			value, ok := data.(*map[string]interface{})
			s.assert.True(ok)
			s.assert.Equal("pong", (*value)["ping"])

			return nil
		}),
	)

	go func() {
		err := messageBroker.Run()
		s.assert.NoError(err)
	}()

	time.Sleep(time.Second * 3)

	// Enviando mensagens para os tópicos inscritos
	for _, topic := range topics {
		messageId, err := s.producer.Publish(topic, message, nil)
		if err != nil {
			received <- true
		}
		s.assert.NoError(err)
		s.assert.NotNil(messageId)
	}

	for i := 0; i < len(topics); i++ {
		<-received
	}
}

func (s *MessageBrokerSubscriberTestSuite) TestFIFOSubscribe() {
	received := make(chan bool, 1)

	subscriberID := "subs-fifo"
	topicID := "topic-test-fifo"

	message := map[string]interface{}{"ping": "pong"}

	go func() {
		messageBroker := grok.NewMessageBrokerSubscriber(
			grok.WithSessionSQS(s.sessionSQS),
			grok.WithSessionSNS(s.sessionSNS),
			grok.WithTopicID(topicID),
			grok.WithSubscriberID(subscriberID),
			grok.WithType(reflect.TypeOf(message)),
			grok.WithFIFO(aws.Bool(true)),
			grok.WithHandler(func(data interface{}) error {
				defer func() { received <- true }()

				value, ok := data.(*map[string]interface{})
				s.assert.True(ok)
				s.assert.Equal("pong", (*value)["ping"])

				return nil
			}),
		)

		err := messageBroker.Run()

		s.assert.NoError(err)
	}()

	time.Sleep(time.Second * 3)

	messageGroupID, err := grok.CreateUuIDV4()
	s.assert.NoError(err)
	s.assert.NotNil(messageGroupID)

	messageDeduplicationID, err := grok.CreateUuIDV4()
	s.assert.NoError(err)
	s.assert.NotNil(messageDeduplicationID)

	messageId, err := s.producer.Publish(topicID, message, grok.WithFIFOAttributes(messageGroupID, messageDeduplicationID))
	if err != nil {
		received <- true
	}

	s.assert.NoError(err)
	s.assert.NotNil(messageId)

	<-received
}
