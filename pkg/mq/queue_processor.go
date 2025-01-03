package mq

import (
	"encoding/json"
	"fmt"
	"sync"
)

func StartQueueProcessor[T any](conn *AMQPConnection, queue, exchange, topic string, processFn func(T) error) {
	var wg sync.WaitGroup

	consume := conn.Consume(queue, WithDefaults(), WithTopicBinding(queue, exchange, topic))

	for i := 0; i < DefaultPrefetch(); i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			var dto T

			for delivery := range consume {
				dto = *new(T)

				if err := json.Unmarshal(delivery.Body, &dto); err != nil {
					fmt.Printf("Error unmarshalling ingress DTO: %v\n", err)
					if err := delivery.Nack(false, false); err != nil {
						fmt.Printf("Error Nacking ingress delivery: %v\n", err)
					}
					continue
				}

				if err := processFn(dto); err != nil {
					fmt.Printf("Error processing ingress DTO: %v\n", err)
					if err := delivery.Nack(false, false); err != nil {
						fmt.Printf("Error Nacking ingress delivery: %v\n", err)
					}
					continue
				}

				if err := delivery.Ack(false); err != nil {
					fmt.Printf("Error Nacking ingress delivery: %v\n", err)
				}
			}
		}(i)
	}

	wg.Wait()
}
