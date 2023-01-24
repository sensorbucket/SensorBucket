package service

import (
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

var _ MessageQueue = (*MessageQueueAMQP)(nil)

type MessageQueueAMQP struct {
	publisher *mq.AMQPPublisher
	exchange  string
}

func NewAMQPQueue(host, exchange string) *MessageQueueAMQP {
	publisher := mq.NewAMQPPublisher(host, exchange, func(c *amqp091.Channel) error {
		return c.ExchangeDeclare(exchange, "topic", true, false, false, false, amqp091.Table{})
	})
	return &MessageQueueAMQP{
		publisher: publisher,
		exchange:  exchange,
	}
}
func (mq *MessageQueueAMQP) Start() {
	mq.publisher.Start()
}
func (mq *MessageQueueAMQP) Shutdown() {
	mq.publisher.Shutdown()
}

func (mq *MessageQueueAMQP) Publish(msg *pipeline.Message) error {
	step, err := msg.NextStep()
	if err != nil {
		return err
	}
	msgData, err := json.Marshal(&msg)
	if err != nil {
		return err
	}
	return mq.publisher.Publish(step, amqp091.Publishing{Body: msgData})
}
