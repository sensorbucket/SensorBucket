package tracing_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
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

func TestRedisStateStorerShouldReturnDefaultValueIfKeyDoesNotExist(t *testing.T) {
	ctx := context.Background()
	redis := createRedis(t)
	stateStore := tracing.NewRedisStore(redis)
	messageID := uuid.NewString()
	step := "latest"

	remainder, err := stateStore.StepsRemainingFor(ctx, messageID, step)
	require.NoError(t, err)

	assert.Equal(t, 9999, remainder)
}

func TestRedisStateStorerShouldUpdateStateAndReturnCorrectValue(t *testing.T) {
	ctx := context.Background()
	messageID := uuid.NewString()
	redis := createRedis(t)
	stateStore := tracing.NewRedisStore(redis)
	remainder := 2
	step := "latest"
	topic := "b"
	ts := time.Now()

	err := stateStore.UpdateState(ctx, messageID, step, remainder, topic, ts)
	gotRemainder, err := stateStore.StepsRemainingFor(ctx, messageID, step)
	require.NoError(t, err)

	assert.Equal(t, remainder, gotRemainder)
}
