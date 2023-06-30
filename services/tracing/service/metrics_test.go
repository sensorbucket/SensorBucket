package tracing_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tracing "sensorbucket.nl/sensorbucket/services/tracing/service"
)

func TestMetrics(t *testing.T) {
	now := time.Now()
	it := &MessageStateIteratorMock{
		NextFunc: func(cursor any) (any, []tracing.MessageState, error) {
			return nil, []tracing.MessageState{
				{
					ID:        uuid.NewString(),
					Timestamp: now,
				},
				{
					ID:        uuid.NewString(),
					Timestamp: now.Add(-15 * time.Second),
				},
				{
					ID:        uuid.NewString(),
					Timestamp: now.Add(-45 * time.Second),
				},
				{
					ID:        uuid.NewString(),
					Timestamp: now.Add(-90 * time.Second),
				},
				{
					ID:        uuid.NewString(),
					Timestamp: now.Add(-230 * time.Second),
				},
				{
					ID:        uuid.NewString(),
					Timestamp: now.Add(-500 * time.Second),
				},
			}, nil
		},
	}
	svc := tracing.NewMetricsService(it)

	metrics, err := svc.Calculate()
	require.NoError(t, err)

	assert.EqualValues(t, tracing.Metrics{
		Count:          6,
		CountBelow30s:  1,
		CountBelow60s:  1,
		CountBelow120s: 1,
		CountBelow300s: 1,
		CountAbove300s: 1,
	}, metrics)
}
