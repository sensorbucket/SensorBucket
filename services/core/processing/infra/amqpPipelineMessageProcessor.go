package processinginfra

import (
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

var _ processing.PipelineMessageProcessor = (*AMQPPipelineMessageProcessor)(nil)

type AMQPPipelineMessageProcessor struct {
	publisher chan<- mq.PublishMessage
}

func NewPipelineMessageProcessor(conn *mq.AMQPConnection, xchg string) *AMQPPipelineMessageProcessor {
	publisher := conn.Publisher(xchg, func(c *amqp091.Channel) error {
		err := c.ExchangeDeclare(xchg, "topic", true, true, false, false, nil)
		return err
	})

	return &AMQPPipelineMessageProcessor{publisher}
}

func (p *AMQPPipelineMessageProcessor) ProcessPipelineMessage(message pipeline.Message) error {
	topic, err := message.NextStep()
	if err != nil {
		return fmt.Errorf("AMQPProcessPipelineMessage could not get message next step: %w", err)
	}
	mqMessage := mq.PublishMessage{
		Topic: topic,
		Publishing: amqp091.Publishing{
			MessageId: message.ID,
			Timestamp: time.Now(),
			Body:      message.Payload,
		},
	}
	p.publisher <- mqMessage
	return nil
}
