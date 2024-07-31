package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/TBD54566975/ftl/internal/observability"
)

const (
	deploymentMeterName = "ftl.deployments.controller"
)

type DeploymentMetrics struct {
	reconciliationFailures metric.Int64Counter
	reconciliationsActive  metric.Int64UpDownCounter
	replicasAdded          metric.Int64Counter
}

func initDeploymentMetrics() (*DeploymentMetrics, error) {
	result := &DeploymentMetrics{}

	var errs error
	var err error

	meter := otel.Meter(deploymentMeterName)

	counter := fmt.Sprintf("%s.reconciliation.failure", deploymentMeterName)
	if result.reconciliationFailures, err = meter.Int64Counter(
		counter,
		metric.WithDescription("the number of failed runner deployment reconciliation tasks")); err != nil {
		result.reconciliationFailures, errs = handleInt64CounterError(counter, err, errs)
	}

	counter = fmt.Sprintf("%s.reconciliations.active", deploymentMeterName)
	if result.reconciliationsActive, err = meter.Int64UpDownCounter(
		counter,
		metric.WithDescription("the number of active deployment reconciliation tasks")); err != nil {
		result.reconciliationsActive, errs = handleInt64UpDownCounterError(counter, err, errs)
	}

	counter = fmt.Sprintf("%s.replicas.added", deploymentMeterName)
	if result.replicasAdded, err = meter.Int64Counter(
		counter,
		metric.WithDescription("the number of runner replicas added (or removed) by the deployment reconciliation tasks")); err != nil {
		result.replicasAdded, errs = handleInt64CounterError(counter, err, errs)
	}

	return result, errs
}

func (m *DeploymentMetrics) ReconciliationFailure(ctx context.Context, module string, key string) {
	m.reconciliationFailures.Add(ctx, 1, metric.WithAttributes(
		attribute.String(observability.ModuleNameAttribute, module),
		attribute.String(observability.RunnerDeploymentKeyAttribute, key),
	))
}

func (m *DeploymentMetrics) ReconciliationStart(ctx context.Context, module string, key string) {
	m.reconciliationsActive.Add(ctx, 1, metric.WithAttributes(
		attribute.String(observability.ModuleNameAttribute, module),
		attribute.String(observability.RunnerDeploymentKeyAttribute, key),
	))
}

func (m *DeploymentMetrics) ReconciliationComplete(ctx context.Context, module string, key string) {
	m.reconciliationsActive.Add(ctx, -1, metric.WithAttributes(
		attribute.String(observability.ModuleNameAttribute, module),
		attribute.String(observability.RunnerDeploymentKeyAttribute, key),
	))
}

func (m *DeploymentMetrics) ReplicasUpdated(ctx context.Context, module string, key string, delta int) {
	m.replicasAdded.Add(ctx, int64(delta), metric.WithAttributes(
		attribute.String(observability.ModuleNameAttribute, module),
		attribute.String(observability.RunnerDeploymentKeyAttribute, key),
	))
}
