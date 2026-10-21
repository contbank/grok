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

// TestConnectWithURLAndAuth garante que o formato novo (URL com senha) funciona sem quebrar o formato
// histórico (host:porta) — parte da migração AWS→Hetzner, onde Redis self-hosted passa a exigir senha.
func (s *RedisTestSuite) TestConnectWithURLAndAuth() {
	s.assert.NotPanics(func() {
		grok.NewRedisConnection("redis://:testpass123@localhost:6380/0")
	})
}

// TestConnectWithURLAndWrongAuth confirma que senha errada continua panicando (mesmo contrato de erro
// de antes: falha de conexão -> panic), não silenciosamente conectando sem autenticar.
func (s *RedisTestSuite) TestConnectWithURLAndWrongAuth() {
	s.assert.Panics(func() {
		grok.NewRedisConnection("redis://:senha-errada@localhost:6380/0")
	})
}
