package zmqbridge

var (
	MESSAGE_CHAN_BACKLOG = 4800
)

// Message is a message to be bridged
type Message struct {
	Content []byte
}
