package measurementtransport

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"

	"sensorbucket.nl/sensorbucket/internal/cleanupper"
	"sensorbucket.nl/sensorbucket/pkg/mq"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	"sensorbucket.nl/sensorbucket/services/core/measurements"
)

func StartMQ(
	svc *measurements.Service,
	conn *mq.AMQPConnection,
	pipelineMessagesExchange,
	measurementQueue,
	measurementStorageTopic,
	measurementErrorTopic string,
	prefetch int,
) cleanupper.Shutdown {
	done := make(chan struct{})
	consume := mq.Consume(conn, measurementQueue, mq.WithDefaults(), mq.WithTopicBinding())
	publish := mq.Publisher(conn, pipelineMessagesExchange, mq.WithDefaults(), mq.WithExchange())

	go func() {
		log.Println("Measurement MQ Transport running...")
		for {
			select {
			case msg := <-consume:
				var pmsg pipeline.Message
				if err := json.Unmarshal(msg.Body, &pmsg); err != nil {
					if nerr := msg.Nack(false, false); nerr != nil {
						err = fmt.Errorf("error nacking message: %w, while handling another error: %w", nerr, err)
					}
					log.Printf("Error unmarshalling amqp message body to pipeline.Message: %v", err)
					continue
				}

				if err := svc.StorePipelineMessage(pmsg); err != nil {
					if nerr := msg.Nack(false, false); nerr != nil {
						err = fmt.Errorf("error nacking message: %w, while handling another error: %w", nerr, err)
					}
					log.Printf("Error storing pipeline message: %v\n", err)
					// Create error
					msgError := pipeline.PipelineError{
						ReceivedByWorker: pmsg,
						Error:            err.Error(),
						Timestamp:        time.Now().UnixMilli(),
						Worker:           "core-measurements",
					}
					msgErrorBytes, err := json.Marshal(msgError)
					if err != nil {
						log.Printf("error marshalling pipeline ErrorMessage into json: %v\n", err)
						continue
					}
					publish <- mq.PublishMessage{
						Topic: measurementErrorTopic,
						Publishing: amqp091.Publishing{
							Body: msgErrorBytes,
						},
					}

					continue
				}
				if err := msg.Ack(false); err != nil {
					log.Printf("Error Acking message: %s\n", err.Error())
				}
			case <-done:
				return
			}
		}
	}()

	return func(ctx context.Context) error {
		close(done)
		return nil
	}
}
