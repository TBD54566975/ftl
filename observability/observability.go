package observability

import (
	"context"

	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Config struct {
	MetricsExporterConfig `embed:"" prefix:"metrics-"`
	SpanExporterConfig    `embed:"" prefix:"traces-"`
}

func Init(ctx context.Context, observabilityServiceClient ftlv1connect.ObservabilityServiceClient, name string, conf Config) {
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(name),
	)

	spanExporter := NewSpanExporter(ctx, observabilityServiceClient, conf.SpanExporterConfig)

	tp := trace.NewTracerProvider(
		trace.WithBatcher(spanExporter),
		trace.WithResource(res))

	otel.SetTracerProvider(tp)

	metricsExporter := NewMetricsExporter(ctx, observabilityServiceClient, conf.MetricsExporterConfig)

	provider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(metric.NewPeriodicReader(metricsExporter, metric.WithInterval(conf.Interval))),
	)

	otel.SetMeterProvider(provider)
}
