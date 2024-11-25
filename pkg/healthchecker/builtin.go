package healthchecker

import "sensorbucket.nl/sensorbucket/pkg/mq"

const (
	MessageQueue = "message_queue"
)

func (b *Builder) WithMessagQueue(conn *mq.AMQPConnection) *Builder {
	check := func() (string, bool) {
		return "state connected", conn.State() == mq.AMQP_CONNECTED
	}
	return b.AddLiveness(MessageQueue, check).AddReadiness(MessageQueue, check)
}
