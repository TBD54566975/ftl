package observability

import (
	"context"
	"errors"
	"fmt"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"strings"
)

const (
	runnerMeterName              = "ftl.runner"
	runnerDeploymentKeyAttribute = "ftl.deployment.key"
	runnerStateNameAttribute     = "ftl.runner.state.name"
)

type RunnerMetrics struct {
	meter                  metric.Meter
	startupFailures        metric.Int64Counter
	registrationHeartbeats metric.Int64Counter
	registrationFailures   metric.Int64Counter
}

func initRunnerMetrics() (*RunnerMetrics, error) {
	result := &RunnerMetrics{}

	var errs error
	var err error

	result.meter = otel.Meter(runnerMeterName)

	counter := fmt.Sprintf("%s.startup.failures", runnerMeterName)
	if result.startupFailures, err = result.meter.Int64Counter(
		counter,
		metric.WithDescription("the number of runner startup failures")); err != nil {
		result.startupFailures, errs = handleInitErrors(counter, err, errs)
	}

	counter = fmt.Sprintf("%s.registration.heartbeats", runnerMeterName)
	if result.registrationHeartbeats, err = result.meter.Int64Counter(
		counter,
		metric.WithDescription("the number of successful runner (re-)registrations")); err != nil {
		result.registrationHeartbeats, errs = handleInitErrors(counter, err, errs)
	}

	counter = fmt.Sprintf("%s.registration.failures", runnerMeterName)
	if result.registrationFailures, err = result.meter.Int64Counter(
		counter,
		metric.WithDescription("the number of failures encounter while attempting to runner registration")); err != nil {
		result.registrationFailures, errs = handleInitErrors(counter, err, errs)
	}

	return result, errs
}

func (m *RunnerMetrics) Registered(ctx context.Context, key *string, state ftlv1.RunnerState) {
	keyAttr := "unknown"
	if key != nil {
		keyAttr = *key
	}
	m.registrationHeartbeats.Add(ctx, 1, metric.WithAttributes(
		attribute.String(runnerDeploymentKeyAttribute, keyAttr),
		attribute.String(runnerStateNameAttribute, runnerStateToString(state)),
	))
}

func (m *RunnerMetrics) RegistrationFailure(ctx context.Context, key *string, state ftlv1.RunnerState) {
	keyAttr := "unknown"
	if key != nil {
		keyAttr = *key
	}
	m.registrationFailures.Add(ctx, 1, metric.WithAttributes(
		attribute.String(runnerDeploymentKeyAttribute, keyAttr),
		attribute.String(runnerStateNameAttribute, runnerStateToString(state)),
	))
}

func (m *RunnerMetrics) StartupFailed(ctx context.Context) {
	m.startupFailures.Add(ctx, 1)
}

//nolint:unparam to suppress noop.Int64Counter{} false complaint
func handleInitErrors(counter string, err error, errs error) (metric.Int64Counter, error) {
	return noop.Int64Counter{}, errors.Join(errs, fmt.Errorf("%q counter init failed; falling back to noop: %w", counter, err))
}

func runnerStateToString(state ftlv1.RunnerState) string {
	return strings.ToLower(strings.TrimPrefix(state.String(), "RUNNER_"))
}
