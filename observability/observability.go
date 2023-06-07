package observability

import (
	"context"
	"net"
	"net/url"
	"time"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	metricsv1 "go.opentelemetry.io/proto/otlp/collector/metrics/v1"

	"github.com/TBD54566975/ftl/internal/3rdparty/protos/opentelemetry/proto/collector/metrics/v1/v1connect"
	"github.com/TBD54566975/ftl/internal/log"
)

type Config struct {
	ObservabilityEndpoint *url.URL      `help:"FTL observability endpoint." env:"FTL_OBSERVABILITY_ENDPOINT" required:""`
	Interval              time.Duration `default:"30s" help:"Interval to export metrics." env:"FTL_METRICS_INTERVAL"`
}

type Observability struct{}

// Export implements v1connect.MetricsServiceHandler
func (*Observability) Export(ctx context.Context, req *connect.Request[metricsv1.ExportMetricsServiceRequest]) (*connect.Response[metricsv1.ExportMetricsServiceResponse], error) {
	logger := log.FromContext(ctx)
	logger.Infof("Metric Req %s:", req.Msg)
	return connect.NewResponse(&metricsv1.ExportMetricsServiceResponse{}), nil
}

func NewService() *Observability {
	return &Observability{}
}

var _ v1connect.MetricsServiceHandler = (*Observability)(nil)

func Init(ctx context.Context, name string, config Config) error {
	_, _, err := net.SplitHostPort(config.ObservabilityEndpoint.Host)
	if err != nil {
		return errors.WithStack(err)
	}

	exporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(config.ObservabilityEndpoint.Host), otlpmetricgrpc.WithInsecure())
	if err != nil {
		return errors.WithStack(err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(name),
	)

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(exporter, metric.WithInterval(config.Interval))),
	)

	otel.SetMeterProvider(meterProvider)

	return nil
}
