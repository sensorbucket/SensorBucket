package mq

import (
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PublishMessage struct {
	Topic      string
	Publishing amqp.Publishing
}

func Publisher(conn *AMQPConnection, xchg string, setup AMQPSetupFunc) chan<- PublishMessage {
	ch := make(chan PublishMessage, 10)
	newConnection := conn.UseConnection()

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
			returns := make(chan amqp.Return)
			amqpChan.NotifyReturn(returns)
			err = setup(amqpChan)
			if err != nil {
				continue
			}

			// Loop until publish channel is closed
			for {
				select {
				case msg, ok := <-returns:
					if !ok {
						continue loopNewConnection
					}
					log.Printf("AMQPPublisher no route to %s (%s)\n", msg.Exchange, msg.RoutingKey)
				case msg, ok := <-ch:
					if !ok {
						continue loopNewConnection
					}
					if msg.Publishing.Headers == nil {
						msg.Publishing.Headers = amqp.Table{}
					}
					msg.Publishing.Headers["timestamp"] = time.Now().UnixMilli()
					amqpChan.Publish(xchg, msg.Topic, true, false, msg.Publishing)
				}
			}
		}
	}()

	return ch
}
