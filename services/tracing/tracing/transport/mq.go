package tracingtransport

import (
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/tracing/tracing"
)

func StartMQ(svc *tracing.Service, conn *mq.AMQPConnection, errQueue string, queue string) {
	pipelineMessages := mq.Consume(conn, queue, setupFunc(queue))

	log.Println("Measurement MQ Transport running")
	go processMessage(queue, pipelineMessages, svc)
}

func processMessage(queue string, deliveries <-chan amqp091.Delivery, svc *tracing.Service) {
	log.Println("Measurement MQ Transport running, tracing pipeline errors...")
	for msg := range deliveries {
		if msg.Timestamp.IsZero() {
			log.Printf("Error: msg timestamp cannot be empty")
			continue
		}
		if msg.RoutingKey == "errors" {
			var res pipeline.PipelineError
			if err := json.Unmarshal(msg.Body, &res); err != nil {
				msg.Nack(false, false)
				log.Printf("Error unmarshalling amqp message body to pipeline.Message: %v", err)
				continue
			}

			if err := svc.HandlePipelineError(res, msg.Timestamp); err != nil {
				msg.Nack(false, false)
				log.Printf("Error handling pipeline message: %v\n", err)
				continue
			}
		} else {
			var res pipeline.Message
			if err := json.Unmarshal(msg.Body, &res); err != nil {
				msg.Nack(false, false)
				log.Printf("Error unmarshalling amqp message body to pipeline.Message: %v", err)
				continue
			}

			if err := svc.HandlePipelineMessage(res, msg.Timestamp); err != nil {
				msg.Nack(false, false)
				log.Printf("Error handling pipeline message: %v\n", err)
				continue
			}
		}
		msg.Ack(false)
	}
}

func setupFunc(queue string) mq.AMQPSetupFunc {
	return func(c *amqp091.Channel) error {
		_, err := c.QueueDeclare(queue, true, false, false, false, nil)
		return err
	}
}
