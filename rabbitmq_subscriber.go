package grok

import (
	"encoding/json"
	"fmt"
	"reflect"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

// runAMQP é o equivalente RabbitMQ do fluxo hoje implementado em checkMessages/createSubscriptionIfNotExists
// (subscriber.go) pra SQS/SNS — mesma assinatura pública (Run() error), backend diferente por dentro. Ver
// /doc/grok-migracao.md §2.1.
func (s *MessageBrokerSubscriber) runAMQP() error {
	if s.dlq && len(s.topicIDs) == 0 {
		return NewError(404, "SUBSCRIBER_ERROR", "error in topic subscriber")
	}

	queueName := s.subscriberID

	if s.dlq {
		if err := s.setupAMQPTopology(); err != nil {
			logrus.WithError(err).Errorf("error starting %s", s.subscriberID)
			return err
		}
	} else {
		// Consumer "cru" de uma fila já existente (ex.: o consumer dedicado da própria DLQ) — mesma
		// semântica do listQueuesBySubscriberID no fluxo SQS: só consome, não provisiona topologia nova.
		// Sem argumentos especiais porque a fila "_dlq" já foi declarada assim por setupAMQPTopology.
		if _, err := s.amqpChannel.QueueDeclare(queueName, true, false, false, false, nil); err != nil {
			logrus.WithError(err).Errorf("error starting %s", s.subscriberID)
			return err
		}
	}

	deliveries, err := s.amqpChannel.Consume(
		queueName,
		"",    // consumer tag (deixa o RabbitMQ gerar)
		false, // autoAck: false — mensagem só sai da fila quando confirmarmos (Ack) explicitamente
		false, // exclusive
		false, // noLocal (RabbitMQ não suporta; sempre false)
		false, // noWait
		nil,   // args
	)
	if err != nil {
		logrus.WithError(err).Errorf("error starting %s", s.subscriberID)
		return err
	}

	logrus.Infof("starting consumer %s", s.subscriberID)

	for delivery := range deliveries {
		s.handleAMQPDelivery(delivery)
	}

	return nil
}

// setupAMQPTopology declara (idempotente) o que o fluxo com DLQ precisa: exchange fanout por tópico
// (auto-provisionamento, igual ao SNS), a fila principal (com dead-letter-exchange apontando pra DLX) e a
// DLX + fila "_dlq" — usando o dead-lettering nativo do RabbitMQ em vez de redrive policy manual.
// Equivalente a createSubscriptionIfNotExists (subscriber.go), mas para RabbitMQ.
func (s *MessageBrokerSubscriber) setupAMQPTopology() error {
	dlxName := fmt.Sprintf("%s_dlx", s.subscriberID)
	dlqName := fmt.Sprintf("%s_dlq", s.subscriberID)

	if err := s.amqpChannel.ExchangeDeclare(dlxName, "fanout", true, false, false, false, nil); err != nil {
		return err
	}

	if _, err := s.amqpChannel.QueueDeclare(dlqName, true, false, false, false, nil); err != nil {
		return err
	}

	if err := s.amqpChannel.QueueBind(dlqName, "", dlxName, false, nil); err != nil {
		return err
	}

	queueArgs := amqp.Table{"x-dead-letter-exchange": dlxName}
	if _, err := s.amqpChannel.QueueDeclare(s.subscriberID, true, false, false, false, queueArgs); err != nil {
		return err
	}

	for _, topicID := range s.topicIDs {
		if err := amqpDeclareTopicExchange(s.amqpChannel, topicID); err != nil {
			return err
		}

		if err := s.amqpChannel.QueueBind(s.subscriberID, "", topicID, false, nil); err != nil {
			return err
		}

		logrus.Infof("starting consumer %s with topic %s", s.subscriberID, topicID)
	}

	return nil
}

// handleAMQPDelivery processa uma mensagem: decodifica, chama o handler do serviço, e decide entre Ack,
// retry (republica na mesma fila com contador incrementado) ou mandar pra DLQ (Nack sem requeue — o
// RabbitMQ roteia automaticamente pra fila "_dlq" via x-dead-letter-exchange da fila principal).
func (s *MessageBrokerSubscriber) handleAMQPDelivery(delivery amqp.Delivery) {
	bodyMessage := reflect.New(s.handleType).Interface()

	if err := json.Unmarshal(delivery.Body, bodyMessage); err != nil {
		logrus.WithError(err).
			WithField("content", string(delivery.Body)).
			Errorf("cannot unmarshal message %s - sending to dlq", delivery.MessageId)

		// Correção intencional em relação ao comportamento antigo do lado SQS (ver /doc/grok-migracao.md
		// §2.1, item 5): lá, mensagem que falhava o unmarshal era apagada da fila principal sem ir pra
		// DLQ, apesar do log dizer o contrário. Aqui o Nack com requeue=false realmente manda pra DLQ.
		delivery.Nack(false, false)
		return
	}

	if err := s.handler(bodyMessage); err != nil {
		s.retryOrDeadLetterAMQP(delivery, err)
		return
	}

	delivery.Ack(false)
}

// retryOrDeadLetterAMQP reproduz o redrive policy do SQS (maxReceiveCount) sem depender de contagem
// automática de entregas — o RabbitMQ não conta tentativas por mensagem como o SQS faz. O contador vai
// num header próprio, incrementado a cada falha; a mensagem original é sempre confirmada (Ack) pra não
// ficar reprocessando em loop apertado — a "nova tentativa" é sempre uma cópia nova republicada.
func (s *MessageBrokerSubscriber) retryOrDeadLetterAMQP(delivery amqp.Delivery, handlerErr error) {
	retryCount := amqpRetryCount(delivery.Headers)

	logrus.WithError(handlerErr).
		Errorf("error handling message %s (tentativa %d/%d)", delivery.MessageId, retryCount+1, s.maxRetries)

	if retryCount >= s.maxRetries-1 {
		// Esgotou as tentativas — Nack sem requeue aciona o dead-letter-exchange da fila principal,
		// que já está configurado (setupAMQPTopology) pra rotear direto pra fila "_dlq".
		delivery.Nack(false, false)
		return
	}

	headers := amqp.Table{}
	for k, v := range delivery.Headers {
		headers[k] = v
	}
	headers[amqpRetryCountHeader] = int32(retryCount + 1)

	err := s.amqpChannel.Publish(
		"",             // exchange padrão
		s.subscriberID, // routing key = nome da fila (truque padrão do RabbitMQ pra publicar direto numa fila)
		false,
		false,
		amqp.Publishing{
			ContentType:  delivery.ContentType,
			DeliveryMode: amqp.Persistent,
			MessageId:    delivery.MessageId,
			Headers:      headers,
			Body:         delivery.Body,
		},
	)

	if err != nil {
		logrus.WithError(err).Errorf("error requeueing message %s for retry", delivery.MessageId)
		// Não conseguiu republicar pra nova tentativa — não perder a mensagem: manda pra DLQ em vez de
		// travar ela indefinidamente sem confirmação.
		delivery.Nack(false, false)
		return
	}

	delivery.Ack(false)
}

// amqpRetryCount lê o header de contagem de tentativas, defensivo quanto ao tipo numérico (o valor pode
// voltar como int32 depois de uma viagem de ida e volta pelo broker).
func amqpRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}

	switch v := headers[amqpRetryCountHeader].(type) {
	case int32:
		return int(v)
	case int64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}
