package ingressarchiver

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/pkg/mq"
)

func StartIngressDTOConsumer(conn *mq.AMQPConnection, svc *Application, queue, xchg, topic string) {
	consume := conn.Consume(queue, func(c *amqp091.Channel) error {
		_, err := c.QueueDeclare(queue, true, false, false, false, nil)
		if err != nil {
			return err
		}

		// Create exchange and bind if both arguments are provided, this is optional
		if xchg != "" && topic != "" {
			if err := c.ExchangeDeclare(xchg, "topic", true, false, false, false, nil); err != nil {
				return err
			}
			if err := c.QueueBind(queue, topic, xchg, false, nil); err != nil {
				return err
			}
		}
		return nil
	})

	for delivery := range consume {
		tracingID := delivery.MessageId
		rawMessage := delivery.Body
		if err := svc.ArchiveIngressDTO(tracingID, rawMessage); err != nil {
			fmt.Printf("Error processing ingress DTO: %v\n", err)
			delivery.Nack(false, false)
			continue
		}
		delivery.Ack(false)
	}
}
