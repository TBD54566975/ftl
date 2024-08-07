package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/TBD54566975/ftl/internal/observability"
)

const (
	deploymentMeterName = "ftl.deployments.controller"
)

type DeploymentMetrics struct {
	reconciliationFailures metric.Int64Counter
	reconciliationsActive  metric.Int64UpDownCounter
	replicasAdded          metric.Int64Counter
	replicasRemoved        metric.Int64Counter
}

func initDeploymentMetrics() (*DeploymentMetrics, error) {
	result := &DeploymentMetrics{
		reconciliationFailures: noop.Int64Counter{},
		reconciliationsActive:  noop.Int64UpDownCounter{},
		replicasAdded:          noop.Int64Counter{},
		replicasRemoved:        noop.Int64Counter{},
	}

	var err error
	meter := otel.Meter(deploymentMeterName)

	signalName := fmt.Sprintf("%s.reconciliation.failures", deploymentMeterName)
	if result.reconciliationFailures, err = meter.Int64Counter(
		signalName,
		metric.WithDescription("the number of failed runner deployment reconciliation tasks")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	signalName = fmt.Sprintf("%s.reconciliations.active", deploymentMeterName)
	if result.reconciliationsActive, err = meter.Int64UpDownCounter(
		signalName,
		metric.WithDescription("the number of active deployment reconciliation tasks")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	signalName = fmt.Sprintf("%s.replicas.added", deploymentMeterName)
	if result.replicasAdded, err = meter.Int64Counter(
		signalName,
		metric.WithDescription("the number of runner replicas added by the deployment reconciliation tasks")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	signalName = fmt.Sprintf("%s.replicas.removed", deploymentMeterName)
	if result.replicasRemoved, err = meter.Int64Counter(
		signalName,
		metric.WithDescription("the number of runner replicas removed by the deployment reconciliation tasks")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	return result, nil
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
	if delta < 0 {
		m.replicasRemoved.Add(ctx, int64(-delta), metric.WithAttributes(
			attribute.String(observability.ModuleNameAttribute, module),
			attribute.String(observability.RunnerDeploymentKeyAttribute, key),
		))
	} else if delta > 0 {
		m.replicasAdded.Add(ctx, int64(delta), metric.WithAttributes(
			attribute.String(observability.ModuleNameAttribute, module),
			attribute.String(observability.RunnerDeploymentKeyAttribute, key),
		))
	}
}
