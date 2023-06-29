package tracing_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sensorbucket.nl/sensorbucket/pkg/pipeline"
	tracing "sensorbucket.nl/sensorbucket/services/tracing/service"
)

func TestServiceShouldArchiveOriginalMessage(t *testing.T) {
	msg := pipeline.NewMessage(uuid.NewString(), []string{"a", "b", "c"})
	msgJSON, err := json.Marshal(&msg)
	require.NoError(t, err)
	delivery := amqp091.Delivery{
		Body:       msgJSON,
		RoutingKey: "first",
		MessageId:  msg.ID,
	}
	archiver := &MessageArchiverMock{
		ArchiveFunc: func(ctx context.Context, delivery amqp091.Delivery) error {
			return nil
		},
	}
	state := &MessageStateStorerMock{
		UpdateStateFunc: func(ctx context.Context, id, key string, stepsRemaining int, currentStep string, timestamp time.Time) error {
			return nil
		},
		StepsRemainingForFunc: func(ctx context.Context, id, step string) (int, error) {
			return 1, nil
		},
	}
	svc := tracing.New(state, archiver)

	err = svc.ProcessDelivery(delivery)
	require.NoError(t, err)

	calls := archiver.ArchiveCalls()
	assert.Len(t, calls, 1)
	assert.Equal(t, delivery.RoutingKey, calls[0].Delivery.RoutingKey)
	assert.Equal(t, delivery.Body, calls[0].Delivery.Body)
	assert.Equal(t, delivery.MessageId, calls[0].Delivery.MessageId)
}

func TestServiceShouldUnsetIfMessageIsFinished(t *testing.T) {
	// TODO: Implement
	t.Fail()
}
