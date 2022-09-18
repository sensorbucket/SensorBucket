package mq

import (
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AMQPConsumer struct {
	amqpHost    string
	queue       string
	consumeChan chan amqp.Delivery

	state         AMQPState
	connectedCond *sync.Cond

	connection *amqp.Connection
	channel    *amqp.Channel

	setupFunc   AMQPSetupFunc
	notifyClose chan *amqp.Error
	done        chan struct{}
}

func NewAMQPConsumer(host, queue string, setupFunc AMQPSetupFunc) *AMQPConsumer {
	return &AMQPConsumer{
		amqpHost:      host,
		queue:         queue,
		done:          make(chan struct{}),
		connectedCond: sync.NewCond(&sync.Mutex{}),
		setupFunc:     setupFunc,
		consumeChan:   make(chan amqp.Delivery, AMQP_QUEUE_LEN),
	}
}

func (a *AMQPConsumer) Start() {
	defer func() {
		log.Println("[AMQPConsumer] Stopping")
		a.state = AMQP_DISCONNECTED
		a.connectedCond.Broadcast()
		a.connection.Close()
	}()

	log.Println("[AMQPConsumer] Starting")
	for {
		a.state = AMQP_RECONNECTING
		log.Println("[AMQPConsumer] Connecting...")
		if err := a.connect(); err != nil {
			log.Printf("[AMQPConsumer] Connection error: %v\n", err)
			select {
			case <-a.done:
				return
			case <-time.After(3 * time.Second):
				continue
			}
		}

		// Connected
		a.state = AMQP_CONNECTED
		a.connectedCond.Broadcast()
		log.Println("[AMQPConsumer] Connected")

		// Start consuming
		consume, err := a.channel.Consume(a.queue, "", false, false, false, false, nil)
		if err != nil {
			select {
			case <-a.done:
				return
			case <-time.After(3 * time.Second):
				goto disconnect
			}
		}

		// Wait for a close or done
		for {
			select {
			case <-a.done:
				return // Returns the `Start` func and executes teardown
			case <-a.notifyClose:
				goto disconnect
			case msg := <-consume:
				a.consumeChan <- msg
			}
		}

	disconnect:
		log.Println("[AMQPConsumer] Disconnected")

		// Prepare for reconnect
		a.connection.Close()
		a.state = AMQP_RECONNECTING
		a.connectedCond.Broadcast()
		select {
		case <-a.done:
			return
		case <-time.After(3 * time.Second):
			continue
		}
	}
}

func (a *AMQPConsumer) Shutdown() {
	close(a.done)
}

func (a *AMQPConsumer) Consume() <-chan amqp.Delivery {
	return a.consumeChan
}

func (a *AMQPConsumer) connect() error {
	conn, err := amqp.Dial(a.amqpHost)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}

	// Setup infrastructure
	if err := a.setupFunc(ch); err != nil {
		return err
	}

	a.notifyClose = ch.NotifyClose(make(chan *amqp.Error))
	a.connection = conn
	a.channel = ch

	return nil
}
