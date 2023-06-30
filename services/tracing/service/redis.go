package tracing

import (
	"context"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

var _ MessageStateStorer = (*RedisStateStore)(nil)
var _ MessageArchiver = (*RedisStateStore)(nil)

type RedisStateStore struct {
	redis      *redis.Client
	archiveTTL time.Duration
	stateTTL   time.Duration
}

func NewRedisStore(client *redis.Client, archiveTTL, stateTTL time.Duration) *RedisStateStore {
	return &RedisStateStore{
		redis:      client,
		archiveTTL: archiveTTL,
		stateTTL:   stateTTL,
	}
}

func (s *RedisStateStore) UpdateState(ctx context.Context, id string, timestamp time.Time) error {
	err := s.redis.HSet(ctx, redisStateKey(id), "id", id, "timestamp", timestamp).Err()
	s.redis.Expire(ctx, redisStateKey(id), s.stateTTL)
	return err
}

func (s *RedisStateStore) FinishState(ctx context.Context, id string) error {
	return s.redis.Del(ctx, redisStateKey(id)).Err()
}

func (s *RedisStateStore) Archive(ctx context.Context, del amqp091.Delivery) error {
	return s.redis.Set(ctx, redisArchiveKey(del.MessageId, del.RoutingKey), string(del.Body), s.archiveTTL).Err()
}

func redisStateKey(id string) string {
	return fmt.Sprintf("messages:%s", id)
}

func redisArchiveKey(id, topic string) string {
	return fmt.Sprintf("messages:%s:topic:%s:archive", id, topic)
}
