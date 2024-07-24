package observability

import (
	"context"
	"fmt"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/observability/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const fsmMeterName = "ftl.fsm"

var fsmMeter = otel.Meter("ftl.fsm")

var fsmCounters = struct {
	instancesActive metric.Int64UpDownCounter
}{}

// TODO error logging and handling
func InitFSMMetrics() {
	fsmCounters.instancesActive, _ = fsmMeter.Int64UpDownCounter(
		fmt.Sprintf("%s.instances.active", fsmMeterName),
		metric.WithDescription("counts the number of active FSM instances"))
}

func FSMInstanceCreated(ctx context.Context, fsm schema.RefKey) {
	fsmCounters.instancesActive.Add(ctx, 1, metric.WithAttributes(
		attribute.String(metrics.ModuleNameAttribute, fsm.Module),
		attribute.String(fsmRefAttribute, fsm.String())))
}

func FSMInstanceCompleted(ctx context.Context, fsm schema.RefKey) {
	fsmCounters.instancesActive.Add(ctx, -1, metric.WithAttributes(
		attribute.String(metrics.ModuleNameAttribute, fsm.Module),
		attribute.String(fsmRefAttribute, fsm.String())))
}
