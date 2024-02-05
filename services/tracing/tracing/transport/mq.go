package tracingtransport

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/tracing/tracing"
)

func StartMQ(svc *tracing.Service, conn *mq.AMQPConnection, queue, xchg, topic string) {
	pipelineMessages := mq.Consume(conn, queue, setupFunc(queue, xchg, topic))

	log.Println("Measurement MQ Transport running")
	go processMessage(pipelineMessages, svc)
}

func processMessage(deliveries <-chan amqp091.Delivery, svc *tracing.Service) {
	log.Println("Measurement MQ Transport running, tracing pipeline errors...")
	for msg := range deliveries {
		tsHeader, ok := msg.Headers["timestamp"]
		if !ok {
			if err := msg.Nack(false, false); err != nil {
				log.Printf("Error: failed to NACK message: %s\n", err.Error())
			}
			log.Printf("Error: Message missing timestamp HEADER\n")
			continue
		}
		tsMilli, ok := tsHeader.(int64)
		if !ok {
			if err := msg.Nack(false, false); err != nil {
				log.Printf("Error: failed to NACK message: %s\n", err.Error())
			}
			log.Printf("Error: Message timestamp header is invalid type: %T\n", tsHeader)
			continue
		}

		ts := time.UnixMilli(tsMilli)
		if ts.IsZero() {
			if err := msg.Nack(false, false); err != nil {
				log.Printf("Error: failed to NACK message: %s\n", err.Error())
			}
			log.Printf("Error: msg timestamp cannot be empty\n")
			continue
		}
		if msg.RoutingKey == "errors" {
			var res pipeline.PipelineError
			if err := json.Unmarshal(msg.Body, &res); err != nil {
				if err := msg.Nack(false, false); err != nil {
					log.Printf("Error: failed to NACK message: %s\n", err.Error())
				}
				log.Printf("Error unmarshalling amqp message body to pipeline.Message: %v", err)
				continue
			}

			if err := svc.HandlePipelineError(res, ts); err != nil {
				if err := msg.Nack(false, false); err != nil {
					log.Printf("Error: failed to NACK message: %s\n", err.Error())
				}
				log.Printf("Error handling pipeline message: %v\n", err)
				continue
			}
		} else {
			var res pipeline.Message
			if err := json.Unmarshal(msg.Body, &res); err != nil {
				if err := msg.Nack(false, false); err != nil {
					log.Printf("Error: failed to NACK message: %s\n", err.Error())
				}
				log.Printf("Error unmarshalling amqp message body to pipeline.Message: %v", err)
				continue
			}

			if err := svc.HandlePipelineMessage(res, ts); err != nil {
				if err := msg.Nack(false, false); err != nil {
					log.Printf("Error: failed to NACK message: %s\n", err.Error())
				}
				log.Printf("Error handling pipeline message: %v\n", err)
				continue
			}
		}
		if err := msg.Ack(false); err != nil {
			log.Printf("Error: failed to ACK message: %s\n", err.Error())
		}
	}
}

func setupFunc(queue, xchg, topic string) mq.AMQPSetupFunc {
	return func(c *amqp091.Channel) error {
		_, err := c.QueueDeclare(queue, true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("error declaring amqp queue: %w", err)
		}
		err = c.ExchangeDeclare(xchg, "topic", true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("error declaring amqp exchange: %w", err)
		}
		err = c.QueueBind(queue, topic, xchg, false, nil)
		if err != nil {
			return fmt.Errorf("error binding amqp queue to exchange: %w", err)
		}
		return nil
	}
}
