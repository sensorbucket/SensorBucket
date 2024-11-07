package processingtransport

import (
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

func StartIngressDTOConsumer(conn *mq.AMQPConnection, svc *processing.Service, queue, xchg, topic string, prefetch int) {
	consume := conn.Consume(queue, func(c *amqp091.Channel) error {
		if err := c.Qos(prefetch, 0, false); err != nil {
			return fmt.Errorf("error setting Qos with prefetch on amqp: %w", err)
		}
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
			if err := delivery.Nack(false, false); err != nil {
				fmt.Printf("Error Nacking ingress delivery: %v\n", err)
			}
			continue
		}

		if err := svc.ProcessIngressDTO(dto); err != nil {
			fmt.Printf("Error processing ingress DTO: %v\n", err)
			if err := delivery.Nack(false, false); err != nil {
				fmt.Printf("Error Nacking ingress delivery: %v\n", err)
			}
			continue
		}
		if err := delivery.Ack(false); err != nil {
			fmt.Printf("Error Nacking ingress delivery: %v\n", err)
		}
	}
}
