package grok_test

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/contbank/grok"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PubSubSubscriberTestSuite struct {
	suite.Suite
	assert     *assert.Assertions
	settings   *grok.Settings
	sessionSQS *session.Session
	sessionSNS *session.Session
	producer   *grok.PubSubProducer
}

func TestPubSubSubscriberTestSuite(t *testing.T) {
	suite.Run(t, new(PubSubSubscriberTestSuite))
}

func (s *PubSubSubscriberTestSuite) SetupTest() {
	s.assert = assert.New(s.T())
	s.settings = &grok.Settings{}
	grok.FromYAML("tests/config.yaml", s.settings)
	s.sessionSQS = grok.CreatePubSubClient(s.settings.AWS) //grok.FakePubSubClient(s.settings.AWS.SQS.Endpoint, s.settings.AWS.SQS.Region)
	s.sessionSNS = grok.CreatePubSubClient(s.settings.AWS) //grok.FakePubSubClient(s.settings.AWS.SNS.Endpoint, s.settings.AWS.SNS.Region)
	s.producer = grok.NewPubSubProducer(s.sessionSNS)

}

func (s *PubSubSubscriberTestSuite) TestSubscribe() {

	subscriberID := "subs"
	topicID := "topic"

	message := map[string]interface{}{"ping": "pong"}

	grok.NewPubSubSubscriber(
		grok.WithSessionSQS(s.sessionSQS),
		grok.WithSessionSNS(s.sessionSNS),
		grok.WithTopicID(topicID),
		grok.WithPubSubSubscriberID(subscriberID),
		grok.WithType(reflect.TypeOf(message)),
		grok.WithHandler(func(data interface{}) error {
			value, ok := data.(*map[string]interface{})
			s.assert.True(ok)
			s.assert.Equal("pong", (*value)["ping"])

			return nil
		}),
	).Run()

	err := s.producer.Publish(topicID, message)
	s.assert.NoError(err)

}
