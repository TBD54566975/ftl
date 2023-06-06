package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

func Int64Counter(ctx context.Context, name string, options ...metric.Int64CounterOption) metric.Int64Counter {
	// TODO(wes): Get the verb name from the context
	counter, err := otel.GetMeterProvider().Meter("verbname").Int64Counter(name, options...)
	if err != nil {
		panic(err)
	}
	return counter
}
