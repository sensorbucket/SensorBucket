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
	AMQP_UNREACHABLE
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
		c.usersLock.Lock()
		for _, user := range c.users {
			close(user)
		}
		c.usersLock.Unlock()
		if c.connection != nil {
			c.state = AMQP_DISCONNECTED
			if err := c.connection.Close(); err != nil {
				log.Printf(
					"AMQPConnection close returned an error but we're treating this as closed anyways: %s\n",
					err.Error(),
				)
			}
		}
		log.Println("AMQPConnection stopped")
	}()

	retries := 0

	// Keep reconnecting until we get a 'done' signal
	for {
		log.Println("AMQPConnection (re)connecting...")
		c.connection = nil
		c.state = AMQP_RECONNECTING
		connection, err := amqp.Dial(c.amqpHost)
		if err != nil {
			// Never give up on reconnecting, the host should act on the not-ready / not-alive state.
			if retries >= c.maximumRetries {
				c.state = AMQP_UNREACHABLE
			}
			backoff := time.Duration(min(retries, c.maximumRetries)) * 3 * time.Second
			log.Printf("AMQPConnection connect failed (attempt %d): %v; retry in %s\n", retries+1, err, backoff)
			select {
			case <-c.done:
				return
			case <-time.After(backoff):
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
			notifyUser(user, c.connection)
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
		if err := c.connection.Close(); err != nil {
			log.Printf(
				"AMQPConnection close returned an error but we're treating this as closed anyways: %s\n",
				err.Error(),
			)
		}
	}
}

func (c *AMQPConnection) State() AMQPState {
	return c.state
}

func (c *AMQPConnection) Shutdown() {
	close(c.done)
}

func (c *AMQPConnection) UseConnection() <-chan *amqp.Connection {
	// Channel must have a 1 buffer, otherwise this gets stuck.
	user := make(chan *amqp.Connection, 1)
	c.usersLock.Lock()
	c.users = append(c.users, user)
	if c.state == AMQP_CONNECTED {
		notifyUser(user, c.connection)
	}
	c.usersLock.Unlock()
	return user
}

// Sends the new connection to the user, while consuming an old one if it's still there.
func notifyUser(user AMQPConnectionUser, conn *amqp.Connection) {
	select {
	case <-user: // drop a stale connection the user hasn't picked up yet
	default:
	}
	user <- conn
}

func (c *AMQPConnection) Consume(queue string, setup ...SetupOption) <-chan amqp.Delivery {
	return Consume(c, queue, setup...)
}

func (c *AMQPConnection) Publisher(xchg string, setup ...SetupOption) chan<- PublishMessage {
	return Publisher(c, xchg, setup...)
}
