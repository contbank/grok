package grok

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

// amqpRetryCountHeader é o header usado pra contar quantas vezes uma mensagem já foi reprocessada, já
// que o RabbitMQ (ao contrário do SQS) não conta tentativas de entrega automaticamente por mensagem.
const amqpRetryCountHeader = "x-retry-count"

// isAMQPEndpoint detecta se uma *session.Session foi criada em modo RabbitMQ (ver CreateSession em
// aws_session.go) — a URL do broker viaja no campo Endpoint da sessão.
func isAMQPEndpoint(s *session.Session) (string, bool) {
	if s == nil || s.Config == nil || s.Config.Endpoint == nil {
		return "", false
	}

	endpoint := *s.Config.Endpoint

	return endpoint, strings.HasPrefix(endpoint, "amqp://") || strings.HasPrefix(endpoint, "amqps://")
}

// amqpDial conecta e abre um canal — falha alta (panic) igual ao padrão já usado em mongo.go/redis.go
// pra erro de conexão de infraestrutura no bootstrap do serviço.
func amqpDial(url string) (*amqp.Connection, *amqp.Channel) {
	conn, err := amqp.Dial(url)
	if err != nil {
		logrus.WithError(err).Panic("error connecting to RabbitMQ")
	}

	ch, err := conn.Channel()
	if err != nil {
		logrus.WithError(err).Panic("error opening RabbitMQ channel")
	}

	return conn, ch
}

// amqpDeclareTopicExchange garante (idempotente) que o "tópico" existe como exchange fanout durável —
// equivalente ao createTopicIfNotExists do SNS (producer.go), com fan-out nativo do RabbitMQ no lugar da
// inscrição SNS->SQS manual.
func amqpDeclareTopicExchange(ch *amqp.Channel, topicID string) error {
	return ch.ExchangeDeclare(
		topicID,  // name
		"fanout", // kind
		true,     // durable
		false,    // auto-delete
		false,    // internal
		false,    // no-wait
		nil,      // args
	)
}

// amqpNewMessageID gera um id de mensagem no mesmo formato usado pelos testes existentes (CreateUuIDV4).
func amqpNewMessageID() string {
	return uuid.New().String()
}
