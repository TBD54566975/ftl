package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/observability"
)

const (
	fsmMeterName             = "ftl.fsm"
	fsmRefAttribute          = "ftl.fsm.ref"
	fsmDestStateRefAttribute = "ftl.fsm.dest.state.ref"
)

type FSMMetrics struct {
	instancesActive    metric.Int64UpDownCounter
	transitionsActive  metric.Int64UpDownCounter
	transitionAttempts metric.Int64Counter
}

func initFSMMetrics() (*FSMMetrics, error) {
	result := &FSMMetrics{
		instancesActive:    noop.Int64UpDownCounter{},
		transitionsActive:  noop.Int64UpDownCounter{},
		transitionAttempts: noop.Int64Counter{},
	}

	var err error
	meter := otel.Meter(fsmMeterName)

	signalName := fmt.Sprintf("%s.instances.active", fsmMeterName)
	if result.instancesActive, err = meter.Int64UpDownCounter(
		signalName,
		metric.WithDescription("counts the number of active FSM instances")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	signalName = fmt.Sprintf("%s.transitions.active", fsmMeterName)
	if result.transitionsActive, err = meter.Int64UpDownCounter(
		signalName,
		metric.WithDescription("counts the number of active FSM transitions")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	signalName = fmt.Sprintf("%s.transitions.attempts", fsmMeterName)
	if result.transitionAttempts, err = meter.Int64Counter(
		signalName,
		metric.WithDescription("counts the number of attempted FSM transitions")); err != nil {
		return nil, wrapErr(signalName, err)
	}

	return result, nil
}

func (m *FSMMetrics) InstanceCreated(ctx context.Context, fsm schema.RefKey) {
	m.instancesActive.Add(ctx, 1, metric.WithAttributes(
		attribute.String(observability.ModuleNameAttribute, fsm.Module),
		attribute.String(fsmRefAttribute, fsm.String())))
}

func (m *FSMMetrics) InstanceCompleted(ctx context.Context, fsm schema.RefKey) {
	m.instancesActive.Add(ctx, -1, metric.WithAttributes(
		attribute.String(observability.ModuleNameAttribute, fsm.Module),
		attribute.String(fsmRefAttribute, fsm.String())))
}

func (m *FSMMetrics) TransitionStarted(ctx context.Context, fsm schema.RefKey, destState schema.RefKey) {
	m.transitionsActive.Add(ctx, 1, metric.WithAttributes(
		attribute.String(observability.ModuleNameAttribute, fsm.Module),
		attribute.String(fsmRefAttribute, fsm.String()),
		attribute.String(fsmDestStateRefAttribute, destState.String())))

	m.transitionAttempts.Add(ctx, 1, metric.WithAttributes(
		attribute.String(observability.ModuleNameAttribute, fsm.Module),
		attribute.String(fsmRefAttribute, fsm.String())))
}

func (m *FSMMetrics) TransitionCompleted(ctx context.Context, fsm schema.RefKey) {
	m.transitionsActive.Add(ctx, -1, metric.WithAttributes(
		attribute.String(observability.ModuleNameAttribute, fsm.Module),
		attribute.String(fsmRefAttribute, fsm.String())))
}
