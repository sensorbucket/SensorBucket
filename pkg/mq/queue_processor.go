package mq

import (
	"errors"
	"fmt"
	"sync"

	"github.com/rabbitmq/amqp091-go"
)

var ErrMalformed = errors.New("delivery malformed")

type (
	ProcessorFunc        = func(delivery amqp091.Delivery) error
	ProcessorFuncBuilder = func() ProcessorFunc
)

// StartQueueProcessor opens a basic consume channel with a queue and exchange topic binding.
// Multiple workers will be started based on the prefetch count. Each worker will call the ProcessFuncBuilder,
// which allows a closure per worker, ie an instantiation of a variable for that worker.
// The processor function will be called for each message received from the queue.
// In case of an error, the message will be requeued unless the error wraps mq.ErrMalformed
// The processFunc parameter is a builder which will be called for
func StartQueueProcessor(conn *AMQPConnection, queue, exchange, topic string, processFunc ProcessorFuncBuilder) {
	var wg sync.WaitGroup

	consume := conn.Consume(queue, WithDefaults(), WithTopicBinding(queue, exchange, topic))

	for i := 0; i < DefaultPrefetch(); i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			process := processFunc()
			for delivery := range consume {
				err := process(delivery)
				if err != nil {
					fmt.Printf("Error: QueueProcessorFunc failed: %s\n", err.Error())
					// Only requeue if err is not an ErrMalformed and it is not already redelivered
					requeue := !errors.Is(err, ErrMalformed) && !delivery.Redelivered
					if err := delivery.Nack(false, requeue); err != nil {
						fmt.Printf("Error: could not NAck delivery: %s\n", err.Error())
					}
					continue
				}

				if err := delivery.Ack(false); err != nil {
					fmt.Printf("Error: could not Ack delivery: %s\n", err.Error())
				}
			}
		}(i)
	}

	wg.Wait()
}
