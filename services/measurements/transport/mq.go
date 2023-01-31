package transport

import (
	"context"
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/measurements/service"
)

type MQTransport struct {
	svc      *service.Service
	consumer *mq.AMQPConsumer
}

func NewMQ(svc *service.Service, consumer *mq.AMQPConsumer) *MQTransport {
	return &MQTransport{
		svc:      svc,
		consumer: consumer,
	}
}

func mqSetupFunc(c *amqp091.Channel) error {
	return nil
}

func (t *MQTransport) Start() {
	for msg := range t.consumer.Consume() {
		var pmsg pipeline.Message
		if err := json.Unmarshal(msg.Body, &pmsg); err != nil {
			msg.Nack(false, false)
			log.Printf("Error unmarshalling amqp message body to pipeline.Message: %v", err)
			return
		}

		if err := t.svc.StorePipelineMessage(context.Background(), pmsg); err != nil {
			msg.Nack(false, false)
			log.Printf("Error storing pipeline message: %v\n", err)
			return
		}

		msg.Ack(false)
	}
}
