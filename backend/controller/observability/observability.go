package observability

import (
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

var (
	AsyncCalls *AsyncCallMetrics
	Deployment *DeploymentMetrics
	FSM        *FSMMetrics
	PubSub     *PubSubMetrics
	Cron       *CronMetrics
)

func init() {
	var errs error
	var err error

	AsyncCalls, err = initAsyncCallMetrics()
	errs = errors.Join(errs, err)
	Deployment, err = initDeploymentMetrics()
	errs = errors.Join(errs, err)
	FSM, err = initFSMMetrics()
	errs = errors.Join(errs, err)
	PubSub, err = initPubSubMetrics()
	errs = errors.Join(errs, err)
	Cron, err = initCronMetrics()
	errs = errors.Join(errs, err)

	if err != nil {
		panic(fmt.Errorf("could not initialize controller metrics: %w", errs))
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

//nolint:unparam
func handleInt64HistogramCounterError(counter string, err error, errs error) (metric.Int64Histogram, error) {
	return noop.Int64Histogram{}, errors.Join(errs, fmt.Errorf("%q counter init failed; falling back to noop: %w", counter, err))
}

func timeSinceMS(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}
