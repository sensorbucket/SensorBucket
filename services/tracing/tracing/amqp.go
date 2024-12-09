package tracing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"sensorbucket.nl/sensorbucket/internal/cleanupper"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

type MQConfig struct {
	Prefetch int

	QueueIngress         string
	ExchangeIngress      string
	ExchangeIngressTopic string

	QueueMessages         string
	ExchangeMessages      string
	ExchangeMessagesTopic string

	ErrorKey string
}

func StartMQTransport(conn *mq.AMQPConnection, svc *Service, config MQConfig) cleanupper.Shutdown {
	if config.ErrorKey == "" {
		config.ErrorKey = "errors"
	}

	stopIngress := ingressConsumer(conn, svc, config)
	stopMessages := messageConsumer(conn, svc, config)
	return func(ctx context.Context) error {
		return errors.Join(
			stopIngress(ctx),
			stopMessages(ctx),
		)
	}
}

func processIngress(svc *Service, delivery *amqp091.Delivery) error {
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

func ingressConsumer(conn *mq.AMQPConnection, svc *Service, config MQConfig) cleanupper.Shutdown {
	done := make(chan struct{})
	incoming := conn.Consume(
		config.QueueIngress,
		mqDefaultPrepare(
			config.Prefetch,
			config.QueueIngress,
			config.ExchangeIngress,
			config.ExchangeIngressTopic,
		),
	)

	go func() {
		for {
			select {
			case <-done:
				return
			case delivery := <-incoming:
				if err := processIngress(svc, &delivery); err != nil {
					delivery.Nack(false, false)
				} else {
					delivery.Ack(false)
				}
			}
		}
	}()

	return func(ctx context.Context) error {
		close(done)
		return nil
	}
}

func processMessage(svc *Service, delivery *amqp091.Delivery, errorKey string) error {
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

func messageConsumer(conn *mq.AMQPConnection, svc *Service, config MQConfig) cleanupper.Shutdown {
	done := make(chan struct{})
	incoming := conn.Consume(
		config.QueueMessages,
		mqDefaultPrepare(
			config.Prefetch,
			config.QueueMessages,
			config.ExchangeMessages,
			config.ExchangeMessagesTopic,
		),
	)

	go func() {
		for {
			select {
			case <-done:
				return
			case delivery := <-incoming:
				if err := processMessage(svc, &delivery, config.ErrorKey); err != nil {
					log.Printf("processing message: %s\n", err.Error())
					delivery.Nack(false, false)
				} else {
					delivery.Ack(false)
				}
			}
		}
	}()

	return func(ctx context.Context) error {
		close(done)
		return nil
	}
}

func mqDefaultPrepare(prefetch int, queue, xchg, topic string) mq.AMQPSetupFunc {
	return func(c *amqp091.Channel) error {
		if err := c.Qos(prefetch, 0, false); err != nil {
			return fmt.Errorf("error setting Qos with prefetch on amqp: %w", err)
		}
		q, err := c.QueueDeclare(queue, true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("error declaring amqp queue: %w", err)
		}
		err = c.ExchangeDeclare(xchg, "topic", true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("error declaring amqp exchange: %w", err)
		}
		err = c.QueueBind(q.Name, topic, xchg, false, nil)
		if err != nil {
			return fmt.Errorf("error binding amqp queue to exchange: %w", err)
		}
		return nil
	}
}
