package timeline

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/block/ftl/internal/observability"
)

const (
	timelineMeterName = "ftl.timeline"
)

type Metrics struct {
	inserted metric.Int64Counter
	dropped  metric.Int64Counter
	failed   metric.Int64Counter
}

var metrics *Metrics

func init() {
	metrics = &Metrics{
		inserted: noop.Int64Counter{},
		dropped:  noop.Int64Counter{},
		failed:   noop.Int64Counter{},
	}

	var err error
	meter := otel.Meter(timelineMeterName)

	signalName := fmt.Sprintf("%s.inserted", timelineMeterName)
	if metrics.inserted, err = meter.Int64Counter(signalName, metric.WithUnit("1"),
		metric.WithDescription("the number of times a timeline event was inserted")); err != nil {
		observability.FatalError(signalName, err)
	}

	signalName = fmt.Sprintf("%s.dropped", timelineMeterName)
	if metrics.dropped, err = meter.Int64Counter(signalName, metric.WithUnit("1"),
		metric.WithDescription("the number of times a timeline event was dropped due to the queue being full")); err != nil {
		observability.FatalError(signalName, err)
	}

	signalName = fmt.Sprintf("%s.failed", timelineMeterName)
	if metrics.dropped, err = meter.Int64Counter(signalName, metric.WithUnit("1"),
		metric.WithDescription("the number of times a timeline event failed to be inserted into the database")); err != nil {
		observability.FatalError(signalName, err)
	}
}

func (m *Metrics) Inserted(ctx context.Context, count int) {
	m.inserted.Add(ctx, int64(count))
}

func (m *Metrics) Dropped(ctx context.Context) {
	m.dropped.Add(ctx, 1)
}

func (m *Metrics) Failed(ctx context.Context, count int) {
	m.failed.Add(ctx, int64(count))
}
