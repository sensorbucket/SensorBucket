package measurements

import (
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

func MQMessageProcessor(svc *Service) mq.ProcessorFuncBuilder {
	return func() mq.ProcessorFunc {
		var msg pipeline.Message
		return func(delivery amqp091.Delivery) error {
			if err := json.Unmarshal(delivery.Body, &msg); err != nil {
				return fmt.Errorf("%w: could not unmarshal delivery body as Pipeline Message: %w", mq.ErrMalformed, err)
			}
			return svc.ProcessPipelineMessage(msg)
		}
	}
}
