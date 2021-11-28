package grok_test

import (
	"testing"

	"github.com/contbank/grok"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type RedisTestSuite struct {
	suite.Suite
	assert   *assert.Assertions
	settings *grok.Settings
}

func TestRedisTestSuite(t *testing.T) {
	suite.Run(t, new(RedisTestSuite))
}

func (s *RedisTestSuite) SetupSuite() {
	s.assert = assert.New(s.T())
	s.settings = &grok.Settings{}
	grok.FromYAML("tests/config.yaml", s.settings)
}

func (s *RedisTestSuite) TestConnect() {
	s.assert.NotPanics(func() {
		grok.NewRedisConnection(s.settings.Redis.ConnectionString)
	})
}

func (s *RedisTestSuite) TestConnectFail() {
	s.assert.Panics(func() {
		grok.NewRedisConnection("nohost")
	})
}
