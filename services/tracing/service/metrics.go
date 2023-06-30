package tracing

import (
	"time"
)

type MessageState struct {
	ID        string
	Timestamp time.Time
}

type MessageStateIterator interface {
	Next(cursor any) (any, []MessageState, error)
}

type MetricsService struct {
	iterator MessageStateIterator
}

func NewMetricsService(iterator MessageStateIterator) *MetricsService {
	return &MetricsService{iterator}
}

type Metrics struct {
	Count          uint
	CountBelow30s  uint
	CountBelow60s  uint
	CountBelow120s uint
	CountBelow300s uint
	CountAbove300s uint
}

func (s *MetricsService) Calculate() (Metrics, error) {
	var metrics Metrics

	var cursor any = nil
	cursor, messages, err := s.iterator.Next(cursor)

	// loop over pages until the end
	for {
		// Update total count
		metrics.Count += uint(len(messages))
		// Loop over messages and calculate metrics
		for _, msg := range messages {
			duration := time.Since(msg.Timestamp)
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
		cursor, messages, err = s.iterator.Next(cursor)
		if err != nil {
			return metrics, err
		}
		if cursor == nil {
			break
		}
	}

	return metrics, nil
}
