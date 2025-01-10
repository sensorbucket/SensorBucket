package tracing

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

func ProcessIngress(svc *Service, delivery *amqp091.Delivery) error {
	queueTime := time.Now()
	hdrTime, hasTime := delivery.Headers["timestamp"]
	if hasTime {
		intTime, ok := hdrTime.(int64)
		if ok {
			queueTime = time.UnixMilli(intTime)
		}
	}

	var msg processing.IngressDTO
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		return fmt.Errorf("unmarshalling IngressDTO message from amqp delivery: %w", err)
	}
	return svc.ProcessTrace(msg, queueTime)
}

func ProcessMessage(svc *Service, delivery *amqp091.Delivery, errorKey string) error {
	queueTime := time.Now()
	hdrTime, hasTime := delivery.Headers["timestamp"]
	if hasTime {
		intTime, ok := hdrTime.(int64)
		if ok {
			queueTime = time.UnixMilli(intTime)
		}
	}
	if delivery.RoutingKey == errorKey {
		var msg pipeline.PipelineError
		if err := json.Unmarshal(delivery.Body, &msg); err != nil {
			return fmt.Errorf("unmarshalling PipelineError message from amqp delivery: %w", err)
		}
		return svc.ProcessTraceError(msg.ReceivedByWorker.TracingID, queueTime, msg.Error)
	}

	var msg pipeline.Message
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		return fmt.Errorf("unmarshalling PipelineMessage message from amqp delivery: %w", err)
	}
	return svc.ProcessTraceStep(msg, queueTime)
}
