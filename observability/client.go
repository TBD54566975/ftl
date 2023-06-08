package observability

import (
	"context"
	"net"

	"github.com/alecthomas/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

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
