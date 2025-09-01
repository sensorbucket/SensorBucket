package measurementsinfra

import (
	"encoding/json"
	"log/slog"

	"github.com/rabbitmq/amqp091-go"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
)

var logger = slog.Default()

func NewStorageErrorPublisher(conn *mq.AMQPConnection, xchg string) measurements.StorageErrorPublisher {
	publisher := conn.Publisher(xchg, func(c *amqp091.Channel) error {
		return c.ExchangeDeclare(xchg, "topic", true, false, false, false, nil)
	})
	messageChan := make(chan *measurements.StorageError, mq.DefaultPrefetch())
	go func() {
		for msg := range messageChan {
			body, err := json.Marshal(msg)
			if err != nil {
				logger.Warn("Could not marshal storage error", "error", msg)
				continue
			}
			publisher <- mq.PublishMessage{
				Topic: "storage_errors",
				Publishing: amqp091.Publishing{
					MessageId: msg.TracingID,
					Body:      body,
				},
			}
		}
	}()

	return messageChan
}
