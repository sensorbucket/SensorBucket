package processing

import (
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/pkg/mq"
)

func MQIngressDTOProcessor(svc *Service) mq.ProcessorFuncBuilder {
	return func() mq.ProcessorFunc {
		var dto IngressDTO
		return func(delivery amqp091.Delivery) error {
			if err := json.Unmarshal(delivery.Body, &dto); err != nil {
				return fmt.Errorf("%w: could not unmarshal delivery body as IngressDTO: %w", mq.ErrMalformed, err)
			}
			return svc.ProcessIngressDTO(dto)
		}
	}
}
