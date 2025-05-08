package measurements

import (
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
)

type StorageErrorPublisher chan<- *StorageError

type StorageError struct {
	TracingID string `json:"tracing_id"`
	Body      []byte `json:"body"`
	Error     string `json:"error"`
}

func MQMessageProcessor(svc *Service, publisher StorageErrorPublisher) mq.ProcessorFuncBuilder {
	return func() mq.ProcessorFunc {
		var msg pipeline.Message
		return func(delivery amqp091.Delivery) error {
			if err := json.Unmarshal(delivery.Body, &msg); err != nil {
				err = fmt.Errorf("%w: could not unmarshal delivery body as Pipeline Message: %w", mq.ErrMalformed, err)
				publisher <- &StorageError{
					TracingID: delivery.MessageId,
					Body:      delivery.Body,
					Error:     err.Error(),
				}
				return err
			}
			return svc.ProcessPipelineMessage(msg)
		}
	}
}
