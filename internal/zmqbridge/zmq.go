package zmqbridge

import (
	"fmt"
	"log"
	"time"

	"github.com/pebbe/zmq4"
)

// ZMQTransport
type ZMQTransport struct {
	done     chan struct{}
	messages chan Message
}

func NewZMQ() *ZMQTransport {
	return &ZMQTransport{}
}

func (z *ZMQTransport) StartConsuming(addr string) error {
	z.done = make(chan struct{})
	z.messages = make(chan Message, MESSAGE_CHAN_BACKLOG)

	ctx, err := zmq4.NewContext()
	if err != nil {
		return fmt.Errorf("error creating ZMQ context: %s", err)
	}
	defer ctx.Term()

	sock, err := ctx.NewSocket(zmq4.ROUTER)
	if err != nil {
		return fmt.Errorf("error creating ZMQ router: %s", err)
	}

	if err := sock.Bind(addr); err != nil {
		return fmt.Errorf("error binding ZMQ router: %s", err)
	}

	// We poll the socket at an interval so the program doesn't block at receiving bytes and we can still shut down gracefully
	poller := zmq4.NewPoller()
	poller.Add(sock, zmq4.POLLIN)

	log.Printf("ZMQ transport active\n")
	defer log.Printf("ZMQ transport stopped\n")

	for {
		select {
		case <-z.done:
			return nil
		default:
			// Wait 50 ms for a message to arrive
			pollSockets, err := poller.Poll(50 * time.Millisecond)
			if err != nil {
				return fmt.Errorf("error polling ZMQ router: %s", err)
			}
			// If no sockets returned then there is nothing to do and we loop
			if len(pollSockets) == 0 {
				continue
			}

			// Otherwise we have a message to process - since we only have 1 socket registered, get the 0th one
			msg, err := pollSockets[0].Socket.RecvMessageBytes(0)
			if err != nil {
				return fmt.Errorf("error receiving message: %s", err)
			}
			// Forward the received message to the message channel
			z.messages <- Message{Content: msg[len(msg)-1]}
		}
	}
}

func (z *ZMQTransport) Channel() <-chan Message {
	return z.messages
}

func (z *ZMQTransport) Shutdown() {
	close(z.done)
	close(z.messages)
}
