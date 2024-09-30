package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const (
	timelineMeterName = "ftl.timeline"
)

type TimelineMetrics struct {
	inserted metric.Int64Counter
	dropped  metric.Int64Counter
	failed   metric.Int64Counter
}

func initTimelineMetrics() (*TimelineMetrics, error) {
	result := &TimelineMetrics{
		inserted: noop.Int64Counter{},
		dropped:  noop.Int64Counter{},
		failed:   noop.Int64Counter{},
	}

	var err error
	meter := otel.Meter(timelineMeterName)

	signalName := fmt.Sprintf("%s.inserted", timelineMeterName)
	if result.inserted, err = meter.Int64Counter(signalName, metric.WithUnit("1"),
		metric.WithDescription("the number of times a timeline event was inserted")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	signalName = fmt.Sprintf("%s.dropped", timelineMeterName)
	if result.dropped, err = meter.Int64Counter(signalName, metric.WithUnit("1"),
		metric.WithDescription("the number of times a timeline event was dropped due to the queue being full")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	signalName = fmt.Sprintf("%s.failed", timelineMeterName)
	if result.dropped, err = meter.Int64Counter(signalName, metric.WithUnit("1"),
		metric.WithDescription("the number of times a timeline event failed to be inserted into the database")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	return result, nil
}

func (m *TimelineMetrics) Inserted(ctx context.Context, count int) {
	m.inserted.Add(ctx, int64(count))
}

func (m *TimelineMetrics) Dropped(ctx context.Context) {
	m.dropped.Add(ctx, 1)
}

func (m *TimelineMetrics) Failed(ctx context.Context, count int) {
	m.failed.Add(ctx, int64(count))
}
