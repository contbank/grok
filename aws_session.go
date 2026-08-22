package grok

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

// CreateSession ...
func CreateSession(settings *AWSCredentials) *session.Session {
	switch {
	case settings.Broker == BrokerKubernetes:
		// Não é AWS de verdade — só um "envelope" pra sinalizar pro consumidor (hoje só
		// LoadSecretsManager) que deve ler de variável de ambiente em vez de chamar a AWS. Mesma
		// estratégia do case BrokerRabbitMQ logo abaixo: preserva o tipo de retorno *session.Session que
		// os consumidores já esperam, sem precisar mudar assinatura de função. Ver secrets_manager.go e
		// /doc/grok-migracao.md.
		return session.Must(session.NewSession(&aws.Config{
			Endpoint: aws.String(BrokerKubernetes),
		}))
	case settings.Broker == BrokerRabbitMQ:
		// RabbitMQ não é AWS — a URL do broker (ex.: amqp://usuario:senha@host:5672/) viaja no mesmo
		// campo Endpoint já usado hoje pra apontar pro Localstack em modo fake. Essa *session.Session é
		// só um "envelope" de transporte pra preservar o tipo de retorno que os ~52 serviços já esperam
		// (`container.SNS`/`container.SQS *session.Session`) — NewMessageBrokerProducer/
		// NewMessageBrokerSubscriber leem `settings.Grok.AWS.SNS/SQS.Endpoint` de volta daqui e nunca
		// falam com a AWS de verdade nesse modo. Ver /doc/grok-migracao.md §2.1.
		return session.Must(session.NewSession(&aws.Config{
			Region:   aws.String(settings.Region),
			Endpoint: aws.String(settings.Endpoint),
		}))
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
