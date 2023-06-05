package observability

import (
	"context"
	"encoding/json"
	"time"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	"github.com/jpillora/backoff"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

var ErrDroppedMetricEvent = errors.New("observability metric event dropped")

type MetricsExporterConfig struct {
	Buffer   int           `default:"1048576" help:"Number of metrics to buffer before dropping."`
	Interval time.Duration `default:"30s" help:"Interval to export metrics." env:"FTL_METRICS_INTERVAL"`
}

type MetricsExporter struct {
	client ftlv1connect.ObservabilityServiceClient
	queue  chan *ftlv1.SendMetricsRequest
}

func NewMetricsExporter(ctx context.Context, client ftlv1connect.ObservabilityServiceClient, config MetricsExporterConfig) *MetricsExporter {
	e := &MetricsExporter{
		client: client,
		queue:  make(chan *ftlv1.SendMetricsRequest, config.Buffer),
	}
	go rpc.RetryStreamingClientStream(ctx, backoff.Backoff{}, e.client.SendMetrics, e.sendLoop)
	return e
}

var _ metric.Exporter = (*MetricsExporter)(nil)

func (e *MetricsExporter) Aggregation(kind metric.InstrumentKind) aggregation.Aggregation {
	return metric.DefaultAggregationSelector(kind)
}

func (e *MetricsExporter) Temporality(kind metric.InstrumentKind) metricdata.Temporality {
	return metric.DefaultTemporalitySelector(kind)
}

func (e *MetricsExporter) Export(ctx context.Context, metrics *metricdata.ResourceMetrics) error {
	data, err := json.Marshal(metrics)
	if err != nil {
		return errors.WithStack(err)
	}
	select {
	case e.queue <- &ftlv1.SendMetricsRequest{Json: data}:
	default:
		return errors.WithStack(ErrDroppedMetricEvent)
	}
	return nil
}

func (e *MetricsExporter) ForceFlush(ctx context.Context) error {
	return nil
}

func (e *MetricsExporter) Shutdown(ctx context.Context) error {
	close(e.queue)
	return nil
}

func (e *MetricsExporter) sendLoop(ctx context.Context, stream *connect.ClientStreamForClient[ftlv1.SendMetricsRequest, ftlv1.SendMetricsResponse]) error {
	logger := log.FromContext(ctx)
	logger.Infof("Metrics send loop started")
	for {
		select {
		case <-ctx.Done():
			return errors.WithStack(context.Cause(ctx))

		case event, ok := <-e.queue:
			if !ok {
				return nil
			}
			logger.Infof("%s", event.Json)
			if err := stream.Send(event); err != nil {
				select {
				case e.queue <- event:
				default:
					log.FromContext(ctx).Errorf(errors.WithStack(ErrDroppedMetricEvent), "metrics queue full while handling error")
				}
				return errors.WithStack(err)
			}
		}
	}
}
