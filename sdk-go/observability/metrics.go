package observability

import (
	"context"

	"github.com/TBD54566975/ftl/internal/rpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

func Int64Counter(ctx context.Context, name string, options ...metric.Int64CounterOption) metric.Int64Counter {
	verb, _ := rpc.VerbFromContext(ctx)
	counter, err := otel.GetMeterProvider().Meter(verb.Name).Int64Counter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Int64UpDownCounter(ctx context.Context, name string, options ...metric.Int64UpDownCounterOption) metric.Int64UpDownCounter {
	verb, _ := rpc.VerbFromContext(ctx)
	counter, err := otel.GetMeterProvider().Meter(verb.Name).Int64UpDownCounter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Int64Histogram(ctx context.Context, name string, options ...metric.Int64HistogramOption) metric.Int64Histogram {
	verb, _ := rpc.VerbFromContext(ctx)
	counter, err := otel.GetMeterProvider().Meter(verb.Name).Int64Histogram(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Int64ObservableCounter(ctx context.Context, name string, options ...metric.Int64ObservableCounterOption) metric.Int64ObservableCounter {
	verb, _ := rpc.VerbFromContext(ctx)
	counter, err := otel.GetMeterProvider().Meter(verb.Name).Int64ObservableCounter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Int64ObservableUpDownCounter(ctx context.Context, name string, options ...metric.Int64ObservableUpDownCounterOption) metric.Int64ObservableUpDownCounter {
	verb, _ := rpc.VerbFromContext(ctx)
	counter, err := otel.GetMeterProvider().Meter(verb.Name).Int64ObservableUpDownCounter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Int64ObservableGauge(ctx context.Context, name string, options ...metric.Int64ObservableGaugeOption) metric.Int64ObservableGauge {
	verb, _ := rpc.VerbFromContext(ctx)
	counter, err := otel.GetMeterProvider().Meter(verb.Name).Int64ObservableGauge(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Float64Counter(ctx context.Context, name string, options ...metric.Float64CounterOption) metric.Float64Counter {
	verb, _ := rpc.VerbFromContext(ctx)
	counter, err := otel.GetMeterProvider().Meter(verb.Name).Float64Counter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Float64UpDownCounter(ctx context.Context, name string, options ...metric.Float64UpDownCounterOption) metric.Float64UpDownCounter {
	verb, _ := rpc.VerbFromContext(ctx)
	counter, err := otel.GetMeterProvider().Meter(verb.Name).Float64UpDownCounter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Float64Histogram(ctx context.Context, name string, options ...metric.Float64HistogramOption) metric.Float64Histogram {
	verb, _ := rpc.VerbFromContext(ctx)
	counter, err := otel.GetMeterProvider().Meter(verb.Name).Float64Histogram(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Float64ObservableCounter(ctx context.Context, name string, options ...metric.Float64ObservableCounterOption) metric.Float64ObservableCounter {
	verb, _ := rpc.VerbFromContext(ctx)
	counter, err := otel.GetMeterProvider().Meter(verb.Name).Float64ObservableCounter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Float64ObservableUpDownCounter(ctx context.Context, name string, options ...metric.Float64ObservableUpDownCounterOption) metric.Float64ObservableUpDownCounter {
	verb, _ := rpc.VerbFromContext(ctx)
	counter, err := otel.GetMeterProvider().Meter(verb.Name).Float64ObservableUpDownCounter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Float64ObservableGauge(ctx context.Context, name string, options ...metric.Float64ObservableGaugeOption) metric.Float64ObservableGauge {
	verb, _ := rpc.VerbFromContext(ctx)
	counter, err := otel.GetMeterProvider().Meter(verb.Name).Float64ObservableGauge(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}
