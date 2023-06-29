package tracing_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	tracing "sensorbucket.nl/sensorbucket/services/tracing/service"
)

func createRedis(t *testing.T) *redis.Client {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "docker.io/redis",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	rc, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "Could not create Redis container")
	t.Cleanup(func() {
		rc.Terminate(ctx)
	})

	port, err := rc.MappedPort(ctx, "6379")
	require.NoError(t, err, "Could not get Redis container port")
	host, err := rc.Host(ctx)
	require.NoError(t, err, "Could not get Redis container host")

	return redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port.Port()),
	})
}

func TestStateManagementAndArchiving(t *testing.T) {
	ctx := context.Background()
	redis := createRedis(t)
	archiveTTL := 15 * time.Minute
	stateTTL := 5 * time.Minute
	stateStore := tracing.NewRedisStore(redis, archiveTTL, stateTTL)
	step := "latest"
	topic := "testtopic"
	now := time.Now()

	t.Run("Steps remaining should default if not set", func(t *testing.T) {
		messageID := uuid.NewString()
		remainder, err := stateStore.StepsRemainingFor(ctx, messageID, step)
		require.NoError(t, err)

		assert.Equal(t, 9999, remainder)
	})

	t.Run("Steps remaining should reflect updated state", func(t *testing.T) {
		messageID := uuid.NewString()
		stepsRemaining := 5
		err := stateStore.UpdateState(ctx, messageID, step, stepsRemaining, topic, now)
		gotRemainder, err := stateStore.StepsRemainingFor(ctx, messageID, step)
		require.NoError(t, err)

		assert.Equal(t, stepsRemaining, gotRemainder)
	})

	t.Run("Should set TTL on archive and state", func(t *testing.T) {
		messageID := uuid.NewString()
		latestStateKey := fmt.Sprintf("messages:%s:step:latest", messageID)
		archiveKey := fmt.Sprintf("messages:%s:topic:%s:archive", messageID, topic)
		delivery := amqp091.Delivery{
			MessageId:  messageID,
			RoutingKey: topic,
			Body:       []byte("testbody"),
		}

		err := stateStore.UpdateState(ctx, messageID, step, 0, topic, now)
		require.NoError(t, err)
		err = stateStore.Archive(ctx, delivery)
		require.NoError(t, err)

		assert.Greater(t, redis.TTL(ctx, archiveKey).Val(), archiveTTL-time.Minute)
		assert.Greater(t, redis.TTL(ctx, latestStateKey).Val(), stateTTL-time.Minute)
	})
}
