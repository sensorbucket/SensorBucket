package tracing

import (
	"encoding/json"
	"fmt"
	"log"
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
		log.Printf("unmarshalling IngressDTO message from amqp delivery: %s\n", err)
	} else {
		if err := svc.StoreTrace(msg, queueTime); err != nil {
			log.Printf("storing trace: %s\n", err)
		}
	}
	if err := svc.StoreIngress(delivery.Body, msg, queueTime); err != nil {
		log.Printf("storing ingress: %s\n", err)
	}

	return nil
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
		return svc.StoreTraceError(msg.ReceivedByWorker.TracingID, queueTime, msg.Error)
	}

	var msg pipeline.Message
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		return fmt.Errorf("unmarshalling PipelineMessage message from amqp delivery: %w", err)
	}
	return svc.StoreTraceStep(msg, queueTime)
}
