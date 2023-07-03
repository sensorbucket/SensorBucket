package tracing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

var _ MessageStateStorer = (*RedisStateStore)(nil)
var _ MessageStateIterator = (*RedisStateStore)(nil)
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

func (s *RedisStateStore) Next(ctx context.Context, cursor any) (any, []MessageState, error) {
	var count int64 = 100
	states := []MessageState{}
	var c uint64
	if cursor != nil {
		var ok bool
		c, ok = cursor.(uint64)
		if !ok {
			return nil, nil, errors.New("Invalid cursor")
		}
	}

	keys, c := s.redis.Scan(ctx, c, redisStateKey("*"), count).Val()
	for _, key := range keys {
		msgID, err := s.redis.HGet(ctx, key, "id").Result()
		if err != nil {
			continue
		}
		ts, err := s.redis.HGet(ctx, key, "timestamp").Time()
		if err != nil {
			continue
		}
		states = append(states, MessageState{
			ID:        msgID,
			Timestamp: ts,
		})
	}

	if c == 0 {
		cursor = nil
	}
	return cursor, states, nil
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
