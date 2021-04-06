package grok

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

// CreateSession ...
func CreateSession(settings *AWSCredentials) *session.Session {
	switch {
	case settings.Fake:
		return FakeSession(settings.Endpoint, settings.Region)
	default:
		sess := session.Must(session.NewSession(&aws.Config{
			Region:      aws.String(settings.Region),
			Credentials: credentials.NewSharedCredentials(settings.Path, "default"),
		}))

		return sess
	}
}

// CreateKMSSession ...
func CreateKMSSession(settings *KMSCredentials) *session.Session {
	return CreateSession(&AWSCredentials{
		Fake:     settings.Fake,
		Path:     settings.Path,
		Endpoint: settings.Endpoint,
		Region:   settings.Region,
	})
}
