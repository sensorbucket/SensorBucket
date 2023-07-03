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
	topic := "testtopic"
	now := time.Now()

	t.Run("Should update state", func(t *testing.T) {
		messageID := uuid.NewString()
		stateKey := fmt.Sprintf("messages:%s", messageID)

		err := stateStore.UpdateState(ctx, messageID, now)
		require.NoError(t, err)

		redisTime, err := redis.HGet(ctx, stateKey, "timestamp").Time()
		require.NoError(t, err)
		assert.WithinDuration(t, now, redisTime, time.Second)

	})

	t.Run("Should set TTL on archive and state", func(t *testing.T) {
		messageID := uuid.NewString()
		stateKey := fmt.Sprintf("messages:%s", messageID)
		archiveKey := fmt.Sprintf("messages:%s:topic:%s:archive", messageID, topic)
		delivery := amqp091.Delivery{
			MessageId:  messageID,
			RoutingKey: topic,
			Body:       []byte("testbody"),
		}

		err := stateStore.UpdateState(ctx, messageID, now)
		require.NoError(t, err)
		err = stateStore.Archive(ctx, delivery)
		require.NoError(t, err)

		assert.Greater(t, redis.TTL(ctx, archiveKey).Val(), archiveTTL-time.Minute)
		assert.Greater(t, redis.TTL(ctx, stateKey).Val(), stateTTL-time.Minute)
	})
}

func TestFinishStateShouldRemoveMessageStateKeys(t *testing.T) {
	ctx := context.Background()
	messageID := uuid.NewString()
	redis := createRedis(t)
	archiveTTL := 15 * time.Minute
	stateTTL := 5 * time.Minute
	stateStore := tracing.NewRedisStore(redis, archiveTTL, stateTTL)
	key := fmt.Sprintf("messages:%s", messageID)

	err := redis.HSet(ctx, key, "timestamp", time.Now()).Err()
	require.NoError(t, err)

	err = stateStore.FinishState(ctx, messageID)
	require.NoError(t, err)

	exists := redis.Exists(ctx, key).Val()
	assert.EqualValues(t, 0, exists)
}

func TestNextShouldIterateOverMessageStates(t *testing.T) {
	ctx := context.Background()
	redis := createRedis(t)
	states := []tracing.MessageState{
		{
			ID:        uuid.NewString(),
			Timestamp: time.Now().Round(time.Minute),
		},
		{
			ID:        uuid.NewString(),
			Timestamp: time.Now().Round(time.Minute),
		},
		{
			ID:        uuid.NewString(),
			Timestamp: time.Now().Round(time.Minute),
		},
		{
			ID:        uuid.NewString(),
			Timestamp: time.Now().Round(time.Minute),
		},
	}
	for _, state := range states {
		err := redis.HSet(ctx, fmt.Sprintf("messages:%s", state.ID), "id", state.ID, "timestamp", state.Timestamp).Err()
		require.NoError(t, err)
	}
	archiveTTL := 15 * time.Minute
	stateTTL := 5 * time.Minute
	store := tracing.NewRedisStore(redis, archiveTTL, stateTTL)

	_, messages, err := store.Next(ctx, nil)
	require.NoError(t, err)

	assert.ElementsMatch(t, states, messages)
}
