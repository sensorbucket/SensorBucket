package mq

import (
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// notifyUser must never block the connection manager, even when the user's
// buffered(1) channel already holds a connection the user hasn't consumed yet.
// Blocking here is what froze the manager (and every other user) during a
// reconnect, leaving connectors alive but not consuming.
func TestNotifyUserDoesNotBlockAndKeepsLatest(t *testing.T) {
	user := make(AMQPConnectionUser, 1)
	first := &amqp.Connection{}
	latest := &amqp.Connection{}

	// Fill the buffer.
	notifyUser(user, first)

	// A second notify with the buffer full must return promptly and replace the
	// stale value rather than block.
	done := make(chan struct{})
	go func() {
		notifyUser(user, latest)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("notifyUser blocked when the user channel buffer was full")
	}

	got := <-user
	if got != latest {
		t.Fatal("expected the user to receive the latest connection, got a stale one")
	}
}
