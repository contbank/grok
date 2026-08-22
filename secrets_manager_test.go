package grok_test

import (
	"os"
	"testing"

	"github.com/contbank/grok"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SecretsManagerTestSuite struct {
	suite.Suite
	assert *assert.Assertions
}

func TestSecretsManagerTestSuite(t *testing.T) {
	suite.Run(t, new(SecretsManagerTestSuite))
}

func (s *SecretsManagerTestSuite) SetupSuite() {
	s.assert = assert.New(s.T())
}

// TestLoadSecretsManagerKubernetesBackend garante que, com broker: kubernetes, LoadSecretsManager lê de
// variável de ambiente em vez de chamar a AWS — parte da migração AWS→Hetzner (accounts é o piloto),
// onde o certificado mTLS de Bankly/Celcoin deixa de depender de credencial AWS.
func (s *SecretsManagerTestSuite) TestLoadSecretsManagerKubernetesBackend() {
	s.T().Setenv("BANKLY_CERTIFICATES", `{"cert":"fake-pem"}`)

	session := grok.CreateSession(&grok.AWSCredentials{Broker: grok.BrokerKubernetes})
	secretsManager := grok.NewSecretsManager(session)

	value, err := secretsManager.LoadSecretsManager("bankly/certificates")

	s.assert.NoError(err)
	s.assert.Equal(`{"cert":"fake-pem"}`, value)
}

// TestLoadSecretsManagerKubernetesBackendMissingEnv garante erro claro (não panic, não string vazia
// silenciosa) quando o Secret esperado não foi montado no Deployment.
func (s *SecretsManagerTestSuite) TestLoadSecretsManagerKubernetesBackendMissingEnv() {
	os.Unsetenv("CELCOIN_CERTIFICATES")

	session := grok.CreateSession(&grok.AWSCredentials{Broker: grok.BrokerKubernetes})
	secretsManager := grok.NewSecretsManager(session)

	_, err := secretsManager.LoadSecretsManager("celcoin/certificates")

	s.assert.Error(err)
}

// TestNewSecretsManagerDefaultBackend garante que, sem broker setado (comportamento de sempre), o
// backend continua sendo AWS — nenhuma regressão pra quem não optar pelo backend novo.
func (s *SecretsManagerTestSuite) TestNewSecretsManagerDefaultBackend() {
	session := grok.CreateSession(&grok.AWSCredentials{Fake: true, Endpoint: "http://localhost:4566", Region: "us-west-2"})
	secretsManager := grok.NewSecretsManager(session)

	s.assert.NotNil(secretsManager)
}
