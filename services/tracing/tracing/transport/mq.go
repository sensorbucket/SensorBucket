package tracingtransport

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/tracing/tracing"
)

func MQMessageProcessor(svc *tracing.Service) mq.ProcessorFuncBuilder {
	return func() mq.ProcessorFunc {
		return func(delivery amqp091.Delivery) error {
			tsHeader, ok := delivery.Headers["timestamp"]
			if !ok {
				return fmt.Errorf("%w: message missing timestamp HEADER", mq.ErrMalformed)
			}
			tsMilli, ok := tsHeader.(int64)
			if !ok {
				return fmt.Errorf("%w: message timestamp header is invalid type: %T", mq.ErrMalformed, tsHeader)
			}

			ts := time.UnixMilli(tsMilli)
			if ts.IsZero() {
				return fmt.Errorf("%w: delivery timestamp cannot be empty", mq.ErrMalformed)
			}
			if delivery.RoutingKey == "errors" {
				var res pipeline.PipelineError
				if err := json.Unmarshal(delivery.Body, &res); err != nil {
					return fmt.Errorf("%w: unmarshalling amqp message body to pipeline.Message: %w", mq.ErrMalformed, err)
				}

				if err := svc.HandlePipelineError(res, ts); err != nil {
					return fmt.Errorf("handling pipeline message: %w", err)
				}
			} else {
				var res pipeline.Message
				if err := json.Unmarshal(delivery.Body, &res); err != nil {
					return fmt.Errorf("%w: unmarshalling amqp message body to pipeline.Message: %w", mq.ErrMalformed, err)
				}

				if err := svc.HandlePipelineMessage(res, ts); err != nil {
					return fmt.Errorf("handling pipeline message: %w", err)
				}
			}
			return nil
		}
	}
}
