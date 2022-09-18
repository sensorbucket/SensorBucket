package mq

import (
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AMQPState uint8

const (
	AMQP_DISCONNECTED AMQPState = iota
	AMQP_CONNECTED
	AMQP_RECONNECTING

	AMQP_QUEUE_LEN = 10
)

type publishMessage struct {
	topic      string
	publishing amqp.Publishing
}

type AMQPSetupFunc func(*amqp.Channel) error

type AMQPPublisher struct {
	amqpHost string
	xchg     string
	pubQueue chan publishMessage

	state         AMQPState
	connectedCond *sync.Cond

	connection *amqp.Connection
	channel    *amqp.Channel

	setupFunc   AMQPSetupFunc
	notifyClose chan *amqp.Error
	done        chan struct{}
}

func NewAMQPPublisher(host, xchg string, setupFunc AMQPSetupFunc) *AMQPPublisher {
	return &AMQPPublisher{
		amqpHost:      host,
		xchg:          xchg,
		pubQueue:      make(chan publishMessage, AMQP_QUEUE_LEN),
		done:          make(chan struct{}),
		connectedCond: sync.NewCond(&sync.Mutex{}),
		setupFunc:     setupFunc,
	}
}

func (a *AMQPPublisher) Start() {
	defer func() {
		log.Println("[AMQPPublisher] Stopping")
		a.state = AMQP_DISCONNECTED
		a.connectedCond.Broadcast()
		a.connection.Close()
	}()

	log.Println("[AMQPPublisher] Starting")
	for {
		a.state = AMQP_RECONNECTING
		log.Println("[AMQPPublisher] Connecting...")
		if err := a.connect(); err != nil {
			log.Printf("[AMQPPublisher] Connection error: %v\n", err)
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
		log.Println("[AMQPPublisher] Connected")

		// Wait for a close or done
	process_loop:
		for {
			select {
			case <-a.done:
				return // Returns the `Start` func and executes teardown
			case <-a.notifyClose:
				break process_loop // Breaks the process_loop causing a reconnect
			case msg := <-a.pubQueue:
				a.channel.Publish(a.xchg, msg.topic, false, false, msg.publishing)
			}
		}
		log.Println("[AMQPPublisher] Disconnected")

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

func (a *AMQPPublisher) Shutdown() {
	close(a.done)
}

func (a *AMQPPublisher) Publish(topic string, publishing amqp.Publishing) error {
	a.pubQueue <- publishMessage{topic, publishing}
	return nil
}

func (a *AMQPPublisher) connect() error {
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
