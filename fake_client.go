package grok

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

// FakeMessageBrokerSession ...
func FakeMessageBrokerSession(endpoint string, region string) *session.Session {
	sess := session.Must(session.NewSession(&aws.Config{
		Region:   aws.String(region),
		Endpoint: aws.String(endpoint),
	}))

	return sess
}
