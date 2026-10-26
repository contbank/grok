package grok

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	amqp "github.com/rabbitmq/amqp091-go"
)

// MaxRequeueDelay é o limite do DelaySeconds do SQS (15 min). Vale também no RabbitMQ, pra os dois backends
// se comportarem igual.
const MaxRequeueDelay = 15 * time.Minute

// RequeueWithDelay republica data só na fila deste subscriber (sem passar pelo tópico, então os outros
// inscritos do tópico não recebem de novo) pra ser entregue depois de delay. Serve pra "tentar de novo mais
// tarde" quando a falha é esperada e passageira (ex.: arquivo ainda não liberado pelo provedor), sem gastar o
// WithMaxRetries do subscriber, que no RabbitMQ republica na hora.
//
// O handler que chama RequeueWithDelay deve retornar nil depois, pra mensagem atual ser confirmada: a nova
// cópia é que chega depois do delay. Chamar de dentro do handler (o canal AMQP é o mesmo do consumo e não é
// seguro entre goroutines).
//
// delay é truncado para segundos e limitado a [0, MaxRequeueDelay]. Filas FIFO do SQS não aceitam delay por
// mensagem e retornam erro.
func (s *MessageBrokerSubscriber) RequeueWithDelay(data interface{}, delay time.Duration) error {
	body, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if delay < 0 {
		delay = 0
	}
	if delay > MaxRequeueDelay {
		delay = MaxRequeueDelay
	}
	seconds := int64(delay / time.Second)

	if s.amqpChannel != nil {
		return s.requeueWithDelayAMQP(body, seconds)
	}

	return s.requeueWithDelaySQS(body, seconds)
}

// requeueWithDelaySQS manda direto na fila do subscriber com DelaySeconds, no mesmo envelope que o SNS entrega
// (checkMessages lê o campo "Message").
func (s *MessageBrokerSubscriber) requeueWithDelaySQS(body []byte, seconds int64) error {
	if s.sqsSvc == nil {
		return NewError(500, "SUBSCRIBER_ERROR", "subscriber without SQS session")
	}

	if s.fifo {
		return NewError(400, "SUBSCRIBER_ERROR", "SQS FIFO queues do not support per-message delay")
	}

	queueURL, err := s.listQueuesBySubscriberID(s.sqsSvc, s.subscriberID)
	if err != nil {
		return err
	}

	if queueURL == nil {
		return NewError(404, "SUBSCRIBER_ERROR", fmt.Sprintf("queue %s not found", s.subscriberID))
	}

	envelope, err := json.Marshal(map[string]string{"Message": string(body)})
	if err != nil {
		return err
	}

	_, err = s.sqsSvc.SendMessage(&sqs.SendMessageInput{
		QueueUrl:     queueURL,
		DelaySeconds: aws.Int64(seconds),
		MessageBody:  aws.String(string(envelope)),
	})

	return err
}

// requeueWithDelayAMQP usa uma fila de espera por valor de delay ("<subscriber>_delay_<N>s"), sem consumidor,
// com TTL fixo: quando a mensagem expira, o dead-lettering nativo do RabbitMQ devolve pra fila do subscriber
// pela exchange padrão. TTL por fila (e não por mensagem) evita que uma mensagem de delay maior segure as de
// delay menor na mesma fila, e não depende do plugin rabbitmq_delayed_message_exchange no broker.
//
// Os argumentos da fila de espera são sempre os mesmos pra um mesmo nome — redeclarar com argumentos
// diferentes fecharia o canal (PRECONDITION_FAILED).
func (s *MessageBrokerSubscriber) requeueWithDelayAMQP(body []byte, seconds int64) error {
	publishing := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    amqpNewMessageID(),
		Body:         body,
	}

	if seconds == 0 {
		return s.amqpChannel.Publish("", s.subscriberID, false, false, publishing)
	}

	delayQueue := fmt.Sprintf("%s_delay_%ds", s.subscriberID, seconds)
	args := amqp.Table{
		"x-message-ttl":             seconds * 1000,
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": s.subscriberID,
	}

	if _, err := s.amqpChannel.QueueDeclare(delayQueue, true, false, false, false, args); err != nil {
		return err
	}

	return s.amqpChannel.Publish("", delayQueue, false, false, publishing)
}
