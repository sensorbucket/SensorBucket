package processingtransport

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

func StartPipelineResultProcessor(svc processing.Service, conn *mq.AMQPConnection, queue string) {
	consumer := conn.Consume(queue, func(c *amqp091.Channel) error {
		_, err := c.QueueDeclare(queue, true, true, false, false, nil)
		return err
	})

	for mqPublishing := range consumer {
		// Convert mqPublishing to pipeline.Message
		var plMessage pipeline.Message
		if err := json.Unmarshal(mqPublishing.Body, &plMessage); err != nil {
			fmt.Printf("PipelineResultProcessor json unmarshal error: %v\n", err)
			continue
		}
		// TODO: How can we be sure this message is correctly formatted?

		ctx := context.TODO()
		err := svc.ProcessPipelineResult(ctx, plMessage)
		if err != nil {
			fmt.Printf("PipelineResultProcessor process result error: %v\n", err)
			requeue := mqPublishing.Redelivered
			mqPublishing.Nack(false, requeue)
			continue
		}

		mqPublishing.Ack(false)
	}
}
