package grok_test

import (
	"testing"

	"github.com/contbank/grok"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MongoTestSuite struct {
	suite.Suite
	assert   *assert.Assertions
	settings *grok.Settings
}

func TestMongoTestSuite(t *testing.T) {
	suite.Run(t, new(MongoTestSuite))
}

func (s *MongoTestSuite) SetupSuite() {
	s.assert = assert.New(s.T())
	s.settings = &grok.Settings{}
	grok.FromYAML("tests/config.yaml", s.settings)
}

func (s *MongoTestSuite) TestConnect() {
	s.assert.NotPanics(func() {
		grok.NewMongoConnection(s.settings.Mongo.ConnectionString, nil)
	})
}

func (s *MongoTestSuite) TestConnectFail() {
	s.assert.Panics(func() {
		grok.NewMongoConnection("nohost", nil)
	})
}

func (s *MongoTestSuite) TestIsNotFoundError() {
	err := grok.NewError(100, "ERROR", "error test message")
	s.assert.False(grok.IsNotFoundError(err))

	err = grok.NewError(100, "ERROR", "account not found")
	s.assert.True(grok.IsNotFoundError(err))

	err = grok.NewError(100, "ERROR", "not found")
	s.assert.True(grok.IsNotFoundError(err))

	err = grok.NewError(100, "SOME_ERROR", "NOT FOUND")
	s.assert.True(grok.IsNotFoundError(err))

	err = grok.NewError(100, "SOME_ERROR", "NO DOCUMENTS IN RESULT")
	s.assert.True(grok.IsNotFoundError(err))

	err = grok.NewError(100, "SOME_ERROR", "no documents in result")
	s.assert.True(grok.IsNotFoundError(err))

	err = grok.NewError(100, "SOME_ERROR", "mongo : no documents in result")
	s.assert.True(grok.IsNotFoundError(err))
}
