package processingtransport

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

func StartIngressDTOConsumer(conn *mq.AMQPConnection, svc *processing.Service, queue, xchg, topic string) {
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
		var dto processing.IngressDTO
		if err := json.Unmarshal(delivery.Body, &dto); err != nil {
			fmt.Printf("Error unmarshalling ingress DTO: %v\n", err)
			delivery.Nack(false, false)
			continue
		}

		if err := svc.ProcessIngressDTO(context.Background(), dto); err != nil {
			fmt.Printf("Error processing ingress DTO: %v\n", err)
			delivery.Nack(false, false)
			continue
		}
		delivery.Ack(false)
	}
}
