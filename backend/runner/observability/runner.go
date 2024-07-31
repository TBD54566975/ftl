package observability

import (
	"context"
	"fmt"
	"strings"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/observability"
	"github.com/alecthomas/types/optional"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	runnerMeterName          = "ftl.runner"
	runnerStateNameAttribute = "ftl.runner.state.name"
)

type RunnerMetrics struct {
	startupFailures        metric.Int64Counter
	registrationHeartbeats metric.Int64Counter
	registrationFailures   metric.Int64Counter
}

func initRunnerMetrics() (*RunnerMetrics, error) {
	result := &RunnerMetrics{}

	var errs error
	var err error

	meter := otel.Meter(runnerMeterName)

	counter := fmt.Sprintf("%s.startup.failures", runnerMeterName)
	if result.startupFailures, err = meter.Int64Counter(
		counter,
		metric.WithDescription("the number of runner startup failures")); err != nil {
		result.startupFailures, errs = handleInt64CounterError(counter, err, errs)
	}

	counter = fmt.Sprintf("%s.registration.heartbeats", runnerMeterName)
	if result.registrationHeartbeats, err = meter.Int64Counter(
		counter,
		metric.WithDescription("the number of successful runner (re-)registrations")); err != nil {
		result.registrationHeartbeats, errs = handleInt64CounterError(counter, err, errs)
	}

	counter = fmt.Sprintf("%s.registration.failures", runnerMeterName)
	if result.registrationFailures, err = meter.Int64Counter(
		counter,
		metric.WithDescription("the number of failures encountered while attempting to register a runner")); err != nil {
		result.registrationFailures, errs = handleInt64CounterError(counter, err, errs)
	}

	return result, errs
}

func (m *RunnerMetrics) Registered(ctx context.Context, key optional.Option[string], state ftlv1.RunnerState) {
	m.registrationHeartbeats.Add(ctx, 1, metric.WithAttributes(
		attribute.String(observability.RunnerDeploymentKeyAttribute, key.Default("unknown")),
		attribute.String(runnerStateNameAttribute, runnerStateToString(state)),
	))
}

func (m *RunnerMetrics) RegistrationFailure(ctx context.Context, key optional.Option[string], state ftlv1.RunnerState) {
	m.registrationFailures.Add(ctx, 1, metric.WithAttributes(
		attribute.String(observability.RunnerDeploymentKeyAttribute, key.Default("unknown")),
		attribute.String(runnerStateNameAttribute, runnerStateToString(state)),
	))
}

func (m *RunnerMetrics) StartupFailed(ctx context.Context) {
	m.startupFailures.Add(ctx, 1)
}

func runnerStateToString(state ftlv1.RunnerState) string {
	return strings.ToLower(strings.TrimPrefix(state.String(), "RUNNER_"))
}
