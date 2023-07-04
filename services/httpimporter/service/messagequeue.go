package service

import (
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

var _ MessageQueue = (*MessageQueueAMQP)(nil)

type MessageQueueAMQP struct {
	connection *mq.AMQPConnection
	publisher  chan<- mq.PublishMessage
}

func NewAMQPQueue(host, exchange string) *MessageQueueAMQP {
	amqpConn := mq.NewConnection(host)
	produceChan := mq.Produce(amqpConn, exchange, func(c *amqp091.Channel) error {
		return nil
	})
	return &MessageQueueAMQP{
		connection: amqpConn,
		publisher:  produceChan,
	}
}
func (m *MessageQueueAMQP) Start() {
	m.connection.Start()
}
func (m *MessageQueueAMQP) Shutdown() {
	m.connection.Shutdown()
}

func (m *MessageQueueAMQP) Publish(msg *pipeline.Message) error {
	step, err := msg.NextStep()
	if err != nil {
		return err
	}
	msgData, err := json.Marshal(&msg)
	if err != nil {
		return err
	}
	m.publisher <- mq.PublishMessage{
		Topic: step,
		Publishing: amqp091.Publishing{
			MessageId: msg.ID,
			Body:      msgData,
		},
	}
	return nil
}
