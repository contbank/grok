package grok

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type SecretsManager struct {
	session    *session.Session
	kubernetes bool
}

// NewSecretsManager ...
//
// Detecta automaticamente se a session recebida é o "envelope" criado por CreateSession quando
// AWSCredentials.Broker == BrokerKubernetes — nesse caso, LoadSecretsManager passa a ler de variável de
// ambiente em vez de chamar a AWS. Nenhum chamador existente precisa mudar: quem não setar
// broker: kubernetes no config.yaml continua batendo na AWS exatamente como antes.
func NewSecretsManager(session *session.Session) *SecretsManager {
	kubernetes := session.Config.Endpoint != nil && *session.Config.Endpoint == BrokerKubernetes
	return &SecretsManager{
		session:    session,
		kubernetes: kubernetes,
	}
}

var secretEnvNameSanitizer = regexp.MustCompile(`[^A-Z0-9]+`)

// secretEnvName deriva o nome da variável de ambiente a partir da chave do segredo — ex.:
// "bankly/certificates" -> "BANKLY_CERTIFICATES". Mesmo critério usado em accounts/pkg/container
// (resolveSecret) pra credencial Intra/KMS, pra manter a convenção de nome previsível entre os dois
// mecanismos.
func secretEnvName(key string) string {
	return strings.Trim(secretEnvNameSanitizer.ReplaceAllString(strings.ToUpper(key), "_"), "_")
}

// LoadSecretsManager ...
func (secretsManager *SecretsManager) LoadSecretsManager(key string) (string, error) {
	if secretsManager.kubernetes {
		envName := secretEnvName(key)
		value := os.Getenv(envName)
		if value == "" {
			return "", fmt.Errorf("secret %q not found: variável de ambiente %s não setada (Secret do Kubernetes ausente ou não montado via envFrom)", key, envName)
		}
		return value, nil
	}

	svc := secretsmanager.New(secretsManager.session)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(key),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := svc.GetSecretValue(input)
	if err != nil {
		return "", err
	}

	return *result.SecretString, nil
}
