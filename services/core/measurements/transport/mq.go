package measurementtransport

import (
	"context"
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
)

func StartMQ(svc *measurements.Service, conn *mq.AMQPConnection, queue string) func() {
	done := make(chan struct{})
	consume := mq.Consume(conn, queue, func(c *amqp091.Channel) error {
		_, err := c.QueueDeclare(queue, true, false, false, false, nil)
		return err
	})

	go func() {
		log.Println("Measurement MQ Transport running...")
		for {
			select {
			case msg := <-consume:
				var pmsg pipeline.Message
				if err := json.Unmarshal(msg.Body, &pmsg); err != nil {
					msg.Nack(false, false)
					log.Printf("Error unmarshalling amqp message body to pipeline.Message: %v", err)
					continue
				}

				if err := svc.StorePipelineMessage(context.Background(), pmsg); err != nil {
					msg.Nack(false, false)
					log.Printf("Error storing pipeline message: %v\n", err)
					continue
				}
				msg.Ack(false)
			case <-done:
				break
			}
		}
	}()

	return func() {
		close(done)
	}
}
