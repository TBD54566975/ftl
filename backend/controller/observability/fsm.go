package observability

import (
	"context"
	"github.com/TBD54566975/ftl/backend/schema"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"sync"
)

var fsmMeter = otel.Meter("ftl.fsm")
var fsmInitOnce = sync.Once{}

var fsmActive metric.Int64UpDownCounter
var fsmTransitions metric.Int64Counter

func initFsmMetrics() {
	fsmInitOnce.Do(func() {
		fsmActive, _ = fsmMeter.Int64UpDownCounter("ftl.fsm.active",
			metric.WithDescription("number of in flight fsm transitions"),
			metric.WithUnit("{count}"))

		fsmTransitions, _ = fsmMeter.Int64Counter("ftl.fsm.transitions",
			metric.WithDescription("number of attempted transitions"),
			metric.WithUnit("{count}"))
	})
}

func RecordFsmTransitionBegin(ctx context.Context, fsm schema.RefKey) {
	initFsmMetrics()

	moduleAttr := metricAttributes.moduleName(fsm.Module)
	featureAttr := metricAttributes.featureName(fsm.Name)

	fsmTransitions.Add(ctx, 1, metric.WithAttributes(moduleAttr, featureAttr))
	fsmActive.Add(ctx, 1, metric.WithAttributes(moduleAttr, featureAttr))
}

func RecordFsmTransitionSuccess(ctx context.Context, fsm schema.RefKey) {
	initFsmMetrics()

	moduleAttr := metricAttributes.moduleName(fsm.Module)
	featureAttr := metricAttributes.featureName(fsm.Name)

	fsmActive.Add(ctx, -1, metric.WithAttributes(moduleAttr, featureAttr))
}
