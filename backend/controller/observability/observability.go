package observability

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/dal"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
)

type metricAttributeBuilders struct {
	moduleName      func(name string) attribute.KeyValue
	featureName     func(name string) attribute.KeyValue
	destinationVerb func(name string) attribute.KeyValue
}

type callMetrics struct {
	meter    metric.Meter
	requests metric.Int64Counter
	failures metric.Int64Counter
	active   metric.Int64UpDownCounter
	latency  metric.Int64Histogram
}

type fsmMetrics struct {
	meter       metric.Meter
	active      metric.Int64UpDownCounter
	transitions metric.Int64Counter
}

type observableMetrics struct {
	attributes metricAttributeBuilders
	calls      *callMetrics
	fsm        *fsmMetrics
}

var (
	metrics = observableMetrics{
		// TODO: move to a initialization method
		attributes: metricAttributeBuilders{
			moduleName: func(name string) attribute.KeyValue {
				return attribute.String("ftl.module.name", name)
			},
			featureName: func(name string) attribute.KeyValue {
				return attribute.String("ftl.feature.name", name)
			},
			destinationVerb: func(name string) attribute.KeyValue {
				return attribute.String("ftl.verb.dest", name)
			},
		},
		calls: &callMetrics{
			meter: otel.Meter("ftl.call"),
		},
		fsm: &fsmMetrics{
			meter: otel.Meter("ftl.fsm"),
		},
	}
)

type CallBegin struct {
	DestVerb *schema.Ref
	Callers  []*schema.Ref
}

type CallEnd struct {
	DeploymentKey model.DeploymentKey
	RequestKey    model.RequestKey
	StartTime     time.Time
	DestVerb      *schema.Ref
	Callers       []*schema.Ref
	Request       *ftlv1.CallRequest
	Response      optional.Option[*ftlv1.CallResponse]
	CallError     optional.Option[error]
}

func init() {
	metrics.calls.requests, _ = metrics.calls.meter.Int64Counter("ftl.call.requests",
		metric.WithDescription("number of verb calls"),
		metric.WithUnit("{count}"))

	metrics.calls.failures, _ = metrics.calls.meter.Int64Counter("ftl.call.failures",
		metric.WithDescription("number of verb call failures"),
		metric.WithUnit("{count}"))

	metrics.calls.active, _ = metrics.calls.meter.Int64UpDownCounter("ftl.call.active",
		metric.WithDescription("number of in flight calls"),
		metric.WithUnit("{count}"))

	metrics.calls.latency, _ = metrics.calls.meter.Int64Histogram("ftl.call.latency",
		metric.WithDescription("verb call latency"),
		metric.WithUnit("{ms}"))

	metrics.fsm.active, _ = metrics.fsm.meter.Int64UpDownCounter("ftl.fsm.active",
		metric.WithDescription("number of in flight fsm transitions"),
		metric.WithUnit("{count}"))

	metrics.fsm.transitions, _ = metrics.fsm.meter.Int64Counter("ftl.fsm.transitions",
		metric.WithDescription("number of attempted transitions"),
		metric.WithUnit("{count}"))
}

func RecordFsmTransitionBegin(ctx context.Context, fsm schema.RefKey) {
	moduleAttr := metrics.attributes.moduleName(fsm.Module)
	featureAttr := metrics.attributes.featureName(fsm.Name)

	metrics.fsm.transitions.Add(ctx, 1, metric.WithAttributes(moduleAttr, featureAttr))
	metrics.fsm.active.Add(ctx, 1, metric.WithAttributes(moduleAttr, featureAttr))
}

func RecordFsmTransitionSuccess(ctx context.Context, fsm schema.RefKey) {
	moduleAttr := metrics.attributes.moduleName(fsm.Module)
	featureAttr := metrics.attributes.featureName(fsm.Name)

	metrics.fsm.active.Add(ctx, -1, metric.WithAttributes(moduleAttr, featureAttr))
}

func RecordCallBegin(ctx context.Context, call *CallBegin) {
	var featureName attribute.KeyValue
	var moduleName attribute.KeyValue
	if len(call.Callers) > 0 {
		featureName = metrics.attributes.featureName(call.Callers[0].Name)
		moduleName = metrics.attributes.moduleName(call.Callers[0].Module)
	} else {
		featureName = metrics.attributes.featureName("unknown")
		moduleName = metrics.attributes.moduleName("unknown")
	}

	destinationVerb := metrics.attributes.destinationVerb(call.DestVerb.Name)

	metrics.calls.active.Add(ctx, 1, metric.WithAttributes(moduleName, featureName, destinationVerb))
}

func RecordCallEnd(ctx context.Context, d *dal.DAL, call *CallEnd) {
	logger := log.FromContext(ctx)
	var sourceVerb optional.Option[schema.Ref]
	var featureName attribute.KeyValue
	var moduleName attribute.KeyValue

	// TODO avoid having to find the source (pass it in `CallEnd` and `CallStart` instead)
	if len(call.Callers) > 0 {
		sourceVerb = optional.Some(*call.Callers[0])
		featureName = metrics.attributes.featureName(call.Callers[0].Name)
		moduleName = metrics.attributes.moduleName(call.Callers[0].Module)
	} else {
		featureName = metrics.attributes.featureName("unknown")
		moduleName = metrics.attributes.moduleName("unknown")
	}

	destinationVerb := metrics.attributes.destinationVerb(call.DestVerb.Name)

	metrics.calls.requests.Add(ctx, 1, metric.WithAttributes(moduleName, featureName, destinationVerb))
	metrics.calls.active.Add(ctx, -1, metric.WithAttributes(moduleName, featureName, destinationVerb))

	var errorStr optional.Option[string]
	var stack optional.Option[string]
	var responseBody []byte

	if callError, ok := call.CallError.Get(); ok {
		errorStr = optional.Some(callError.Error())
		metrics.calls.failures.Add(ctx, 1, metric.WithAttributes(moduleName, featureName, destinationVerb))
	} else if response, ok := call.Response.Get(); ok {
		responseBody = response.GetBody()
		if callError := response.GetError(); callError != nil {
			errorStr = optional.Some(callError.Message)
			stack = optional.Ptr(callError.Stack)
			metrics.calls.failures.Add(ctx, 1, metric.WithAttributes(moduleName, featureName, destinationVerb))
		}
	}

	metrics.calls.latency.Record(ctx, time.Now().Sub(call.StartTime).Milliseconds(), metric.WithAttributes(moduleName, featureName, destinationVerb))

	err := d.InsertCallEvent(ctx, &dal.CallEvent{
		Time:          call.StartTime,
		DeploymentKey: call.DeploymentKey,
		RequestKey:    optional.Some(call.RequestKey),
		Duration:      time.Since(call.StartTime),
		SourceVerb:    sourceVerb,
		DestVerb:      *call.DestVerb,
		Request:       call.Request.GetBody(),
		Response:      responseBody,
		Error:         errorStr,
		Stack:         stack,
	})
	if err != nil {
		logger.Errorf(err, "failed to record call")
	}
}
