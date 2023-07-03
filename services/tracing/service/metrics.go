package tracing

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var _ prometheus.Collector = (*MetricsService)(nil)

type MessageState struct {
	ID        string
	Timestamp time.Time
}

type MessageStateIterator interface {
	Next(context context.Context, cursor any) (any, []MessageState, error)
}

type MetricsService struct {
	iterator MessageStateIterator
}

func NewMetricsService(iterator MessageStateIterator) *MetricsService {
	svc := &MetricsService{
		iterator: iterator,
	}
	return svc
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
	ctx := context.TODO()
	var metrics Metrics

	var cursor any = nil
	cursor, messages, err := s.iterator.Next(ctx, cursor)

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
		cursor, messages, err = s.iterator.Next(ctx, cursor)
		if err != nil {
			return metrics, err
		}
		if cursor == nil {
			break
		}
	}

	return metrics, nil
}

func (s *MetricsService) Collect(ch chan<- prometheus.Metric) {
	metr, err := s.Calculate()
	if err != nil {
		ch <- prometheus.NewInvalidMetric(countDesc, fmt.Errorf("could not calculate metrics: %w", err))
		return
	}
	ch <- prometheus.MustNewConstMetric(
		countDesc,
		prometheus.GaugeValue,
		float64(metr.Count),
		"all",
	)
	ch <- prometheus.MustNewConstMetric(
		countDesc,
		prometheus.GaugeValue,
		float64(metr.CountBelow30s),
		"30s",
	)
	ch <- prometheus.MustNewConstMetric(
		countDesc,
		prometheus.GaugeValue,
		float64(metr.CountBelow60s),
		"60s",
	)
	ch <- prometheus.MustNewConstMetric(
		countDesc,
		prometheus.GaugeValue,
		float64(metr.CountBelow120s),
		"120s",
	)
	ch <- prometheus.MustNewConstMetric(
		countDesc,
		prometheus.GaugeValue,
		float64(metr.CountBelow300s),
		"300s",
	)
	ch <- prometheus.MustNewConstMetric(
		countDesc,
		prometheus.GaugeValue,
		float64(metr.CountAbove300s),
		"300s+",
	)
}

var (
	countDesc = prometheus.NewDesc(
		"sensorbucket_messages_in_pipeline_count",
		"The amount of messages currently in the pipeline by time spent in current step",
		[]string{"duration"}, nil,
	)
)

func (s *MetricsService) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(s, ch)
}
