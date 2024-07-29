package observability

import (
	"context"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const (
	fsmMeterName             = "ftl.runner"
	fsmRefAttribute          = "ftl.fsm.ref"
	fsmDestStateRefAttribute = "ftl.fsm.dest.state.ref"
)

type RunnerMetrics struct {
	meter           metric.Meter
	startupFailures metric.Int64Counter
	instancesActive metric.Int64UpDownCounter
}

func initRunnerMetrics() (*RunnerMetrics, error) {
	result := &RunnerMetrics{}

	var errs error
	var err error

	result.meter = otel.Meter(fsmMeterName)

	counter := fmt.Sprintf("%s.startup.Failures", fsmMeterName)
	if result.startupFailures, err = result.meter.Int64Counter(
		counter,
		metric.WithDescription("counts the number of runner startup failures")); err != nil {
		errs = joinInitErrors(counter, err, errs)
		result.startupFailures = noop.Int64Counter{}
	}

	counter = fmt.Sprintf("%s.instances.active", fsmMeterName)
	if result.instancesActive, err = result.meter.Int64UpDownCounter(
		counter,
		metric.WithDescription("counts the number of active runner instances")); err != nil {
		errs = joinInitErrors(counter, err, errs)
		result.instancesActive = noop.Int64UpDownCounter{}
	}

	return result, errs
}

func (m *RunnerMetrics) RunnerStarted(ctx context.Context) {
	m.instancesActive.Add(ctx, 1, metric.WithAttributes())
}

func (m *RunnerMetrics) RunnerStartupFailed(ctx context.Context) {
	m.startupFailures.Add(ctx, 1, metric.WithAttributes())
}

func joinInitErrors(counter string, err error, errs error) error {
	return errors.Join(errs, fmt.Errorf("%q counter init failed; falling back to noop: %w", counter, err))
}
