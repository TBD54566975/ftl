package observability

import (
	"context"
	"fmt"

	"github.com/alecthomas/types/optional"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/TBD54566975/ftl/internal/observability"
)

const (
	deploymentMeterName = "ftl.deployments.runner"
)

type DeploymentMetrics struct {
	failure metric.Int64Counter
	active  metric.Int64UpDownCounter
}

func initDeploymentMetrics() (*DeploymentMetrics, error) {
	result := &DeploymentMetrics{
		failure: noop.Int64Counter{},
		active:  noop.Int64UpDownCounter{},
	}

	var err error
	meter := otel.Meter(deploymentMeterName)

	counter := fmt.Sprintf("%s.failures", deploymentMeterName)
	if result.failure, err = meter.Int64Counter(
		counter,
		metric.WithDescription("the number of deployment failures")); err != nil {
		return nil, wrapErr(counter, err)
	}

	counter = fmt.Sprintf("%s.active", deploymentMeterName)
	if result.active, err = meter.Int64UpDownCounter(
		counter,
		metric.WithDescription("the number of active deployments")); err != nil {
		return nil, wrapErr(counter, err)
	}

	return result, nil
}

func (m *DeploymentMetrics) Failure(ctx context.Context, key optional.Option[string]) {
	m.failure.Add(ctx, 1, metric.WithAttributes(
		attribute.String(observability.RunnerDeploymentKeyAttribute, key.Default("unknown")),
	))
}

func (m *DeploymentMetrics) Started(ctx context.Context, key string) {
	m.active.Add(ctx, 1, metric.WithAttributes(
		attribute.String(observability.RunnerDeploymentKeyAttribute, key),
	))
}

func (m *DeploymentMetrics) Completed(ctx context.Context, key string) {
	m.active.Add(ctx, -1, metric.WithAttributes(
		attribute.String(observability.RunnerDeploymentKeyAttribute, key),
	))
}
