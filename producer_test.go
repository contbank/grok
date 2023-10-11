package grok_test

import (
	"testing"

	"github.com/contbank/grok"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ProducerTestSuite struct {
	suite.Suite
	assert   *assert.Assertions
	settings *grok.Settings
}

func TestProducerTestSuite(t *testing.T) {
	suite.Run(t, new(ProducerTestSuite))
}

func (s *ProducerTestSuite) SetupTest() {
	s.assert = assert.New(s.T())
	s.settings = &grok.Settings{}
	grok.FromYAML("tests/config.yaml", s.settings)
}

func (s *ProducerTestSuite) TestPublish() {
	session := grok.FakeSession(s.settings.AWS.SNS.Endpoint, s.settings.AWS.SNS.Region)
	producer := grok.NewMessageBrokerProducer(session)

	messageId, err := producer.Publish("test-topic", map[string]interface{}{"ping": "pong"}, nil)

	s.assert.NoError(err)
	s.assert.NotNil(messageId)
}

func (s *ProducerTestSuite) TestPublishFIFO() {
	session := grok.FakeSession(s.settings.AWS.SNS.Endpoint, s.settings.AWS.SNS.Region)
	producer := grok.NewMessageBrokerProducer(session)

	messageGroupID, err := grok.CreateUuIDV4()
	s.assert.NoError(err)
	s.assert.NotNil(messageGroupID)

	messageDeduplicationID, err := grok.CreateUuIDV4()
	s.assert.NoError(err)
	s.assert.NotNil(messageDeduplicationID)

	messageId, err := producer.Publish(
		"test-topic-fifo",
		map[string]interface{}{"ping": "pong"},
		grok.WithFIFOAttributes(messageGroupID, messageDeduplicationID),
	)

	s.assert.NoError(err)
	s.assert.NotNil(messageId)
}
