package processingtransport

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/pkg/ingress"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

func StartIngressProcessingListener(svc processing.Service, conn *mq.AMQPConnection, queue string) {
	consumer := conn.Consume(queue, func(c *amqp091.Channel) error {
		_, err := c.QueueDeclare(queue, true, true, false, false, nil)
		return err
	})

	for mqPublishing := range consumer {
		// Convert mqPublishing to IngressDTO
		var ingressDTO ingress.DTO
		if err := json.Unmarshal(mqPublishing.Body, &ingressDTO); err != nil {
			fmt.Printf("IngressProcessingListener json unmarshal error: %v\n", err)
			continue
		}
		if err := ingressDTO.Validate(); err != nil {
			fmt.Printf("IngressProcessingListener validation error: %v\n", err)
			continue
		}

		ctx := context.TODO()
		err := svc.ProcessIngressDTO(ctx, ingressDTO)
		if err != nil {
			fmt.Printf("IngressProcessingListener process ingress data error: %v\n", err)
			requeue := mqPublishing.Redelivered
			mqPublishing.Nack(false, requeue)
			continue
		}

		mqPublishing.Ack(false)
	}
}
