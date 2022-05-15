package zmqbridge

import (
	"errors"
	"log"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

var (
	ErrAMQPShutdown = errors.New("AMQP connection is shutdown")
	RETRY_DELAY     = 5 * time.Second
)

// AMQPTransport is a publisher connection to the amqp server
type AMQPTransport struct {
	conn *amqp091.Connection
	xchg string

	notifyClose chan *amqp091.Error
	done        chan struct{}
	messages    chan Message
}

func NewAMQP(xchg string) *AMQPTransport {
	return &AMQPTransport{
		xchg:     xchg,
		messages: make(chan Message, MESSAGE_CHAN_BACKLOG),
	}
}

func (t *AMQPTransport) Channel() chan<- Message {
	return t.messages
}

func (t *AMQPTransport) Start(addr string) error {
	t.done = make(chan struct{})

	log.Printf("Starting AMQP transport for %s\n", addr)
	defer log.Println("Stopped AMQP transport")

	t.done = make(chan struct{})
	for {
		if err := t.connect(addr); err != nil {
			log.Printf("Error connecting to AMQP server: %s\n", err)
			if retry(RETRY_DELAY, t.done) {
				continue
			}
			return ErrAMQPShutdown
		}

		// Connection succesful
		log.Printf("Connected to AMQP server\n")

		// Create channel
		ch, err := t.conn.Channel()
		if err != nil {
			log.Printf("Error creating AMQP channel: %s\n", err)
			if retry(RETRY_DELAY, t.done) {
				continue
			}
			return ErrAMQPShutdown
		}

		// Declare exchange if not exists
		if err := ch.ExchangeDeclare(t.xchg, "topic", true, false, false, false, nil); err != nil {
			log.Printf("Error creating Exchange: %s\n", err)
			if retry(RETRY_DELAY, t.done) {
				continue
			}
			return ErrAMQPShutdown
		}

		// Wait for shutdown or close
	process_loop:
		for {
			select {
			case msg := <-t.messages:
				if err := ch.Publish(t.xchg, "", false, false, amqp091.Publishing{
					Body: msg.Content,
				}); err != nil {
					log.Printf("Error publishing message: %s\n", err)
				} else {
					log.Printf("Published message\n")
				}
			case <-t.notifyClose:
				log.Printf("AMQP connection closed\n")
				break process_loop
			case <-t.done:
				return ErrAMQPShutdown
			}
		}
	}
}

func retry(delay time.Duration, done chan struct{}) bool {
	select {
	case <-done:
		return false
	case <-time.After(delay):
		return true
	}
}

func (t *AMQPTransport) connect(addr string) error {
	if t.conn != nil {
		t.conn.Close()
	}

	conn, err := amqp091.Dial(addr)
	if err != nil {
		return err
	}

	t.notifyClose = conn.NotifyClose(make(chan *amqp091.Error))
	t.conn = conn

	return nil
}

func (t *AMQPTransport) Shutdown() {
	close(t.done)
}
