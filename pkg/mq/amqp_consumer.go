package mq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

func Consume(conn *AMQPConnection, queue string, opts ...SetupOption) <-chan amqp.Delivery {
	ch := make(chan amqp.Delivery, DefaultPrefetch())
	newConnection := conn.UseConnection()

	go func() {
	loopNewConnection:
		for {
			amqpConn, ok := <-newConnection
			if !ok {
				return
			}
			amqpChan, err := amqpConn.Channel()
			if err != nil {
				continue
			}
			err = setupChannel(amqpChan, opts)
			if err != nil {
				continue
			}

			amqpDeliveryChan, err := amqpChan.Consume(queue, "", false, false, false, false, nil)
			if err != nil {
				continue
			}

			// Loop until delivery channel is closed
			for {
				msg, ok := <-amqpDeliveryChan
				if !ok {
					continue loopNewConnection
				}
				ch <- msg
			}
		}
	}()

	return ch
}
