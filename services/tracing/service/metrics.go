package tracing

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type MetricsService struct {
	redis redis.Client
}

type Metrics struct {
	Count          uint
	CountBelow30s  uint
	CountBelow60s  uint
	CountBelow120s uint
	CountBelow300s uint
	CountAbove300s uint
}

func (s *MetricsService) calculate() (Metrics, error) {
	ctx := context.Background()
	var metrics Metrics

	pattern := "messages:*:step:latest"
	var cursor uint64 = 0
	var pageSize int64 = 100
	keys, cursor, err := s.redis.Scan(ctx, cursor, pattern, pageSize).Result()
	if err != nil {
		return metrics, err
	}

	// loop over pages until the end
	for cursor != 0 {
		// Update total count
		metrics.Count += uint(len(keys))
		// Loop over keys and calculate metrics
		for _, key := range keys {
			//id, err := s.redis.HGet(ctx, key, "id").Result()
			//if err != nil {
			//	fmt.Printf("could not get id for key '%s': %v\n", key, err)
			//    continue
			//}
			since, err := s.redis.HGet(ctx, key, "timestamp").Time()
			if err != nil {
				fmt.Printf("could not get timestamp for key '%s': %v\n", key, err)
				continue
			}
			duration := time.Since(since)
			switch {
			case duration < 5*time.Second:
				// Do nothing
			case duration < 30*time.Second:
				metrics.CountBelow30s++
			case duration < 60*time.Second:
				metrics.CountBelow60s++
			case duration < 120*time.Second:
				metrics.CountBelow120s++
			case duration < 300*time.Second:
				metrics.CountBelow300s++
			default:
				metrics.CountAbove300s++
			}
		}

		// Get next page
		keys, cursor, err = s.redis.Scan(ctx, cursor, pattern, pageSize).Result()
		if err != nil {
			return metrics, err
		}
	}

	return metrics, nil
}
