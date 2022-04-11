package grok

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

type SecretsManager struct {
	session *session.Session
}

func NewSecretsManager(session *session.Session) *SecretsManager {
	return &SecretsManager{
		session: session,
	}
}

func (secretsManager *SecretsManager) LoadSecretsManager(key string) (string, error) {
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
