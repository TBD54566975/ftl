package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"

	"github.com/TBD54566975/ftl/internal/rpc"
)

func MeterWithVerb(ctx context.Context) metric.Meter {
	verb, ok := rpc.VerbFromContext(ctx)
	if !ok {
		panic("traces: no verb in context")
	}
	return otel.GetMeterProvider().Meter(verb.Name)
}

func Int64Counter(ctx context.Context, name string, options ...metric.Int64CounterOption) metric.Int64Counter {
	counter, err := MeterWithVerb(ctx).Int64Counter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Int64UpDownCounter(ctx context.Context, name string, options ...metric.Int64UpDownCounterOption) metric.Int64UpDownCounter {
	counter, err := MeterWithVerb(ctx).Int64UpDownCounter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Int64Histogram(ctx context.Context, name string, options ...metric.Int64HistogramOption) metric.Int64Histogram {
	counter, err := MeterWithVerb(ctx).Int64Histogram(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Int64ObservableCounter(ctx context.Context, name string, options ...metric.Int64ObservableCounterOption) metric.Int64ObservableCounter {
	counter, err := MeterWithVerb(ctx).Int64ObservableCounter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Int64ObservableUpDownCounter(ctx context.Context, name string, options ...metric.Int64ObservableUpDownCounterOption) metric.Int64ObservableUpDownCounter {
	counter, err := MeterWithVerb(ctx).Int64ObservableUpDownCounter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Int64ObservableGauge(ctx context.Context, name string, options ...metric.Int64ObservableGaugeOption) metric.Int64ObservableGauge {
	counter, err := MeterWithVerb(ctx).Int64ObservableGauge(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Float64Counter(ctx context.Context, name string, options ...metric.Float64CounterOption) metric.Float64Counter {
	counter, err := MeterWithVerb(ctx).Float64Counter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Float64UpDownCounter(ctx context.Context, name string, options ...metric.Float64UpDownCounterOption) metric.Float64UpDownCounter {
	counter, err := MeterWithVerb(ctx).Float64UpDownCounter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Float64Histogram(ctx context.Context, name string, options ...metric.Float64HistogramOption) metric.Float64Histogram {
	counter, err := MeterWithVerb(ctx).Float64Histogram(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Float64ObservableCounter(ctx context.Context, name string, options ...metric.Float64ObservableCounterOption) metric.Float64ObservableCounter {
	counter, err := MeterWithVerb(ctx).Float64ObservableCounter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Float64ObservableUpDownCounter(ctx context.Context, name string, options ...metric.Float64ObservableUpDownCounterOption) metric.Float64ObservableUpDownCounter {
	counter, err := MeterWithVerb(ctx).Float64ObservableUpDownCounter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}

func Float64ObservableGauge(ctx context.Context, name string, options ...metric.Float64ObservableGaugeOption) metric.Float64ObservableGauge {
	counter, err := MeterWithVerb(ctx).Float64ObservableGauge(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}
