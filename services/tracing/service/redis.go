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

func (s *RedisStateStore) StepsRemainingFor(ctx context.Context, id, step string) (int, error) {
	r := s.redis
	ok, err := r.Exists(ctx, redisMessageKey(id, step)).Result()
	if err != nil {
		return 0, err
	}
	if ok == 0 {
		return 9999, nil
	}
	remainder, err := r.HGet(ctx, redisMessageKey(id, step), "remainder").Int()
	return remainder, err
}

func (s *RedisStateStore) UpdateState(ctx context.Context, id, step string, remainder int, topic string, timestamp time.Time) error {
	err := s.redis.HSet(ctx, redisMessageKey(id, step), "id", id, "remainder", remainder, "topic", topic, "timestamp", timestamp).Err()
	s.redis.Expire(ctx, redisMessageKey(id, step), s.stateTTL)
	return err
}

func (s *RedisStateStore) FinishState(ctx context.Context, id string) error {
	return s.redis.Del(ctx, redisMessageKey(id, "latest")).Err()
}

func (s *RedisStateStore) Archive(ctx context.Context, del amqp091.Delivery) error {
	return s.redis.Set(ctx, redisArchiveKey(del.MessageId, del.RoutingKey), string(del.Body), s.archiveTTL).Err()
}

func redisMessageKey(id, step string) string {
	return fmt.Sprintf("messages:%s:step:%s", id, step)
}

func redisArchiveKey(id, topic string) string {
	return fmt.Sprintf("messages:%s:topic:%s:archive", id, topic)
}
