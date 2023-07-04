package mq

import (
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AMQPSetupFunc func(*amqp.Channel) error

type AMQPState uint8

const (
	AMQP_DISCONNECTED AMQPState = iota
	AMQP_CONNECTED
	AMQP_RECONNECTING

	AMQP_QUEUE_LEN = 10
)

type AMQPConnectionUser chan *amqp.Connection

type AMQPConnection struct {
	amqpHost    string
	state       AMQPState
	notifyClose chan *amqp.Error
	done        chan struct{}

	connection     *amqp.Connection
	usersLock      sync.Mutex
	users          []AMQPConnectionUser
	maximumRetries int
}

func NewConnection(host string) *AMQPConnection {
	conn := &AMQPConnection{
		amqpHost:    host,
		state:       AMQP_DISCONNECTED,
		notifyClose: make(chan *amqp.Error),
		done:        make(chan struct{}),

		usersLock:      sync.Mutex{},
		users:          make([]AMQPConnectionUser, 0),
		maximumRetries: 10,
	}
	return conn
}

func (c *AMQPConnection) Start() {
	defer func() {
		log.Println("AMQPConnection stopping")
		c.state = AMQP_DISCONNECTED
		c.usersLock.Lock()
		for _, user := range c.users {
			close(user)
		}
		c.usersLock.Unlock()
		c.connection.Close()
		log.Println("AMQPConnection stopped")
	}()

	retries := 0
	// Keep reconnecting until we get a 'done' signal
	for {
		log.Println("AMQPConnection (re)connecting...")
		c.state = AMQP_RECONNECTING
		connection, err := amqp.Dial(c.amqpHost)
		if err != nil {
			log.Printf("AMQPConnection connect failed: %v\n", err)
			if retries > c.maximumRetries {
				log.Printf("AMQPConnection maximum retries of %d reached, quitting...\n", retries)
				return
			}
			log.Printf("AMQPConnection retry in %d seconds...\n", retries*3)
			select {
			case <-c.done:
				return
			case <-time.After(time.Duration(retries) * time.Second * 3):
				retries++
				continue
			}
		}
		retries = 0
		c.connection = connection
		c.notifyClose = connection.NotifyClose(make(chan *amqp.Error))
		log.Printf("AMQPConnection connection succes\n")

		// Notify connection users of new connection
		c.usersLock.Lock()
		for _, user := range c.users {
			user <- c.connection
		}
		c.state = AMQP_CONNECTED
		c.usersLock.Unlock()

		// Wait for done or disconnect
		select {
		case <-c.done:
			return
		case <-c.notifyClose:
			// Continue
		}

		// Disconnected, so close to be sure
		log.Printf("AMQPConnection disconnected\n")
		c.connection.Close()
	}
}

func (c *AMQPConnection) Shutdown() {
	close(c.done)
}

func (c *AMQPConnection) UseConnection() <-chan *amqp.Connection {
	user := make(chan *amqp.Connection)
	c.usersLock.Lock()
	c.users = append(c.users, user)
	if c.state == AMQP_CONNECTED {
		user <- c.connection
	}
	c.usersLock.Unlock()
	return user
}
