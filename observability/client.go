package observability

import (
	"context"
	"net"
	"net/url"
	"time"

	"github.com/alecthomas/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Config struct {
	Endpoint *url.URL      `help:"FTL observability endpoint." env:"FTL_OBSERVABILITY_ENDPOINT" placeholder:"URL"`
	Interval time.Duration `default:"30s" help:"Interval to export metrics." env:"FTL_METRICS_INTERVAL"`
}

func Init(ctx context.Context, name string, config Config) error {
	if config.Endpoint == nil {
		return errors.Errorf("FTL_OBSERVABILITY_ENDPOINT is required")
	}
	_, _, err := net.SplitHostPort(config.Endpoint.Host)
	if err != nil {
		return errors.WithStack(err)
	}

	otelLogger := NewOtelLogger(ctx)
	otel.SetLogger(otelLogger)
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		otelLogger.Error(err, "OTEL")
	}))

	ftlExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(config.Endpoint.Host), otlpmetricgrpc.WithInsecure())
	if err != nil {
		return errors.WithStack(err)
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(name),
	)

	otelExporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(ftlExporter, metric.WithInterval(config.Interval))),
		metric.WithReader(metric.NewPeriodicReader(otelExporter, metric.WithInterval(config.Interval))),
	)

	otel.SetMeterProvider(meterProvider)

	return nil
}
