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
		cred := &credentials.SharedCredentialsProvider{
			Profile: "default",
		}

		if len(settings.Profile) > 0 {
			cred.Profile = settings.Profile
		}

		if len(settings.Path) > 0 {
			cred.Filename = settings.Path
		}

		sess := session.Must(session.NewSession(&aws.Config{
			Region:      aws.String(settings.Region),
			Credentials: credentials.NewCredentials(cred),
		}))

		return sess
	}
}
