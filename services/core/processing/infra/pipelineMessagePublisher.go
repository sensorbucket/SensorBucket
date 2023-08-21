package processinginfra

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

var pipelineMessagePublisherBuffer = 10

func NewPipelineMessagePublisher(conn *mq.AMQPConnection, xchg string) processing.PipelineMessagePublisher {
	publisher := conn.Publisher(xchg, func(c *amqp091.Channel) error {
		return c.ExchangeDeclare(xchg, "topic", true, false, false, false, nil)
	})
	messageChan := make(chan *pipeline.Message, pipelineMessagePublisherBuffer)
	go func() {
		for msg := range messageChan {
			topic, err := msg.CurrentStep()
			if err != nil {
				fmt.Printf("PipelineMessagePublisher could not get step from pipeline message: %v\n", err)
				continue
			}
			log.Printf("Publishing to topic %s\n", topic)
			jsonData, err := json.Marshal(msg)
			if err != nil {
				fmt.Printf("PipelineMessagePublisher could not marshal pipeline message: %v", err)
				continue
			}
			publishing := mq.PublishMessage{
				Topic: topic,
				Publishing: amqp091.Publishing{
					MessageId: msg.ID,
					Body:      jsonData,
				},
			}
			publisher <- publishing
		}
	}()

	return messageChan
}
