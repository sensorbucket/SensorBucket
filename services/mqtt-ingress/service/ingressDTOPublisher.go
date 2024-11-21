package service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/metric"

	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/services/core/processing"
)

var buffer = 1000

func StartIngressDTOPublisher(conn *mq.AMQPConnection, xchg, topic string) chan<- processing.IngressDTO {
	publisher := conn.Publisher(xchg, func(c *amqp091.Channel) error {
		return c.ExchangeDeclare(xchg, "topic", true, false, false, false, nil)
	})
	dtoC := make(chan processing.IngressDTO, buffer)
	_, err := meter.Int64ObservableGauge("dto_publisher_backlog_count", metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
		o.Observe(int64(len(dtoC)))
		return nil
	}))
	if err != nil {
		log.Printf("In IngressDTOPublisher, could not create backlog metric: %s\n", err.Error())
	}
	go func() {
		log.Println("IngressDTOPublisher running...")
		for dto := range dtoC {
			jsonData, err := json.Marshal(dto)
			if err != nil {
				log.Printf("IngressDTOPublisher error marshalling dto: %v\n", err)
				continue
			}
			publisher <- mq.PublishMessage{
				Topic: topic,
				Publishing: amqp091.Publishing{
					MessageId: dto.TracingID.String(),
					Body:      jsonData,
				},
			}
		}
	}()
	return dtoC
}
