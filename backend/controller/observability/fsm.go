package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/observability"
)

const (
	fsmMeterName    = "ftl.fsm"
	fsmRefAttribute = "ftl.fsm.ref"
)

var fsmMeter = otel.Meter("ftl.fsm")

var fsmCounters = struct {
	instancesActive metric.Int64UpDownCounter
}{}

func InitFSMMetrics() error {
	var err error

	fsmCounters.instancesActive, err = fsmMeter.Int64UpDownCounter(
		fmt.Sprintf("%s.instances.active", fsmMeterName),
		metric.WithDescription("counts the number of active FSM instances"))

	if err != nil {
		return fmt.Errorf("could not initialize fsm metrics: %w", err)
	}

	return nil
}

func FSMInstanceCreated(ctx context.Context, fsm schema.RefKey) {
	if fsmCounters.instancesActive != nil {
		fsmCounters.instancesActive.Add(ctx, 1, metric.WithAttributes(
			attribute.String(observability.ModuleNameAttribute, fsm.Module),
			attribute.String(fsmRefAttribute, fsm.String())))
	}
}

func FSMInstanceCompleted(ctx context.Context, fsm schema.RefKey) {
	if fsmCounters.instancesActive != nil {
		fsmCounters.instancesActive.Add(ctx, -1, metric.WithAttributes(
			attribute.String(observability.ModuleNameAttribute, fsm.Module),
			attribute.String(fsmRefAttribute, fsm.String())))
	}
}
