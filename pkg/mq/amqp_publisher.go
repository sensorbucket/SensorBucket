package mq

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type publishMessage struct {
	topic      string
	publishing amqp.Publishing
}

func Produce(conn *AMQPConnection, xchg string, setup AMQPSetupFunc) chan<- publishMessage {
	ch := make(chan publishMessage, 10)
	newConnection := conn.UseConnection()
	returns := make(chan amqp.Return)

	go func() {
	loopNewConnection:
		for {
			amqpConn, ok := <-newConnection
			if !ok {
				log.Println("AMQPPublisher lost connection")
				return
			}
			amqpChan, err := amqpConn.Channel()
			if err != nil {
				continue
			}
			amqpChan.NotifyReturn(returns)
			err = setup(amqpChan)
			if err != nil {
				continue
			}

			// Loop until publish channel is closed
			for {
				select {
				case msg := <-returns:
					log.Printf("AMQPPublisher no route to %s (%s)\n", msg.Exchange, msg.RoutingKey)
				case msg, ok := <-ch:
					if !ok {
						continue loopNewConnection
					}
					amqpChan.Publish(xchg, msg.topic, true, false, msg.publishing)
				}
			}
		}
	}()

	return ch
}
