package grok

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

// CreateMessageBrokerSession ...
func CreateMessageBrokerSession(settings *AWSCredentials) *session.Session {
	switch {
	case settings.Fake:
		return FakeMessageBrokerSession(settings.Endpoint, settings.Region)
	default:
		sess := session.Must(session.NewSession(&aws.Config{
			Region:      aws.String(settings.Region),
			Credentials: credentials.NewSharedCredentials(settings.Path, "default"),
		}))

		return sess
	}
}
