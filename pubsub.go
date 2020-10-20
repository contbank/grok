package grok

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

// CreatePubSubClient ...
func CreatePubSubClient(settings *AWSSettings) *session.Session {
	switch {
	case settings.SNS.Fake:
		return FakePubSubClient(settings.SNS.Endpoint, settings.SNS.Region)
	case settings.SQS.Fake:
		return FakePubSubClient(settings.SQS.Endpoint, settings.SQS.Region)
	default:
		sess := session.Must(session.NewSession(&aws.Config{
			Region:      aws.String(settings.SNS.Region),
			Credentials: credentials.NewSharedCredentials(settings.Filename, "default"),
		}))

		return sess
	}
}
