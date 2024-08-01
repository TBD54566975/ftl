package observability

import (
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

var (
	Runner     *RunnerMetrics
	Deployment *DeploymentMetrics
)

func init() {
	var errs error
	var err error

	Runner, err = initRunnerMetrics()
	errs = errors.Join(errs, err)
	Deployment, err = initDeploymentMetrics()
	errs = errors.Join(errs, err)

	if errs != nil {
		panic(fmt.Errorf("could not initialize runner metrics: %w", err))
	}
}

//nolint:unparam
func handleInt64CounterError(counter string, err error, errs error) (metric.Int64Counter, error) {
	return noop.Int64Counter{}, errors.Join(errs, fmt.Errorf("%q counter init failed; falling back to noop: %w", counter, err))
}

//nolint:unparam
func handleInt64UpDownCounterError(counter string, err error, errs error) (metric.Int64UpDownCounter, error) {
	return noop.Int64UpDownCounter{}, errors.Join(errs, fmt.Errorf("%q counter init failed; falling back to noop: %w", counter, err))
}
