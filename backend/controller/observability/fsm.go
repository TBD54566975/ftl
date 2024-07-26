package observability

import (
	"context"
	"errors"
	"fmt"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
)

const (
	fsmMeterName             = "ftl.fsm"
	fsmRefAttribute          = "ftl.fsm.ref"
	fsmDestStateRefAttribute = "ftl.fsm.dest.state.ref"
)

type FSMMetrics struct {
	meter              metric.Meter
	instancesActive    metric.Int64UpDownCounter
	transitionsActive  metric.Int64UpDownCounter
	transitionAttempts metric.Int64Counter
}

func initFSMMetrics() (*FSMMetrics, error) {
	result := &FSMMetrics{}

	var errs error
	var err error

	result.meter = otel.Meter("ftl.fsm")

	counter := fmt.Sprintf("%s.instances.active", fsmMeterName)
	if result.instancesActive, err = result.meter.Int64UpDownCounter(
		counter,
		metric.WithDescription("counts the number of active FSM instances")); err != nil {
		errs = errors.Join(errs, fmt.Errorf("%q counter init failed; falling back to noop: %w", counter, err))
		result.instancesActive = noop.Int64UpDownCounter{}
	}

	counter = fmt.Sprintf("%s.transitions.active", fsmMeterName)
	if result.transitionsActive, err = result.meter.Int64UpDownCounter(
		counter,
		metric.WithDescription("counts the number of active FSM transitions")); err != nil {
		errs = errors.Join(errs, fmt.Errorf("%q counter init failed; falling back to noop: %w", counter, err))
		result.transitionsActive = noop.Int64UpDownCounter{}
	}

	counter = fmt.Sprintf("%s.transitions.attempts", fsmMeterName)
	if result.transitionAttempts, err = result.meter.Int64Counter(
		counter,
		metric.WithDescription("counts the number of attempted FSM transitions")); err != nil {
		errs = errors.Join(errs, fmt.Errorf("%q counter init failed; falling back to noop: %w", counter, err))
		result.transitionAttempts = noop.Int64Counter{}
	}

	return result, errs
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
