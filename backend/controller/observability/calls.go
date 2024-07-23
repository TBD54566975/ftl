package observability

import (
	"context"
	"github.com/TBD54566975/ftl/backend/controller/dal"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/alecthomas/types/optional"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"sync"
	"time"
)

var callMeter = otel.Meter("ftl.call")
var callInitOnce = sync.Once{}

var callRequests metric.Int64Counter
var callFailures metric.Int64Counter
var callActive metric.Int64UpDownCounter
var callLatency metric.Int64Histogram

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

func initCallMetrics() {
	callInitOnce.Do(func() {
		callRequests, _ = callMeter.Int64Counter("ftl.call.requests",
			metric.WithDescription("number of verb calls"),
			metric.WithUnit("{count}"))
		callFailures, _ = callMeter.Int64Counter("ftl.call.failures",
			metric.WithDescription("number of verb call failures"),
			metric.WithUnit("{count}"))
		callActive, _ = callMeter.Int64UpDownCounter("ftl.call.active",
			metric.WithDescription("number of in flight calls"),
			metric.WithUnit("{count}"))
		callLatency, _ = callMeter.Int64Histogram("ftl.call.latency",
			metric.WithDescription("verb call latency"),
			metric.WithUnit("{ms}"))
	})
}

func RecordCallBegin(ctx context.Context, call *CallBegin) {
	initCallMetrics()

	var featureName attribute.KeyValue
	var moduleName attribute.KeyValue
	if len(call.Callers) > 0 {
		featureName = metricAttributes.featureName(call.Callers[0].Name)
		moduleName = metricAttributes.moduleName(call.Callers[0].Module)
	} else {
		featureName = metricAttributes.featureName("unknown")
		moduleName = metricAttributes.moduleName("unknown")
	}

	destinationVerb := metricAttributes.destinationVerb(call.DestVerb.Name)

	callActive.Add(ctx, 1, metric.WithAttributes(moduleName, featureName, destinationVerb))
}

func RecordCallEnd(ctx context.Context, d *dal.DAL, call *CallEnd) {
	initCallMetrics()

	logger := log.FromContext(ctx)
	var sourceVerb optional.Option[schema.Ref]
	var featureName attribute.KeyValue
	var moduleName attribute.KeyValue

	// TODO avoid having to find the source (pass it in `CallEnd` and `CallStart` instead)
	if len(call.Callers) > 0 {
		sourceVerb = optional.Some(*call.Callers[0])
		featureName = metricAttributes.featureName(call.Callers[0].Name)
		moduleName = metricAttributes.moduleName(call.Callers[0].Module)
	} else {
		featureName = metricAttributes.featureName("unknown")
		moduleName = metricAttributes.moduleName("unknown")
	}

	destinationVerb := metricAttributes.destinationVerb(call.DestVerb.Name)

	callRequests.Add(ctx, 1, metric.WithAttributes(moduleName, featureName, destinationVerb))
	callActive.Add(ctx, -1, metric.WithAttributes(moduleName, featureName, destinationVerb))

	var errorStr optional.Option[string]
	var stack optional.Option[string]
	var responseBody []byte

	if callError, ok := call.CallError.Get(); ok {
		errorStr = optional.Some(callError.Error())
		callFailures.Add(ctx, 1, metric.WithAttributes(moduleName, featureName, destinationVerb))
	} else if response, ok := call.Response.Get(); ok {
		responseBody = response.GetBody()
		if callError := response.GetError(); callError != nil {
			errorStr = optional.Some(callError.Message)
			stack = optional.Ptr(callError.Stack)
			callFailures.Add(ctx, 1, metric.WithAttributes(moduleName, featureName, destinationVerb))
		}
	}

	callLatency.Record(ctx, time.Now().Sub(call.StartTime).Milliseconds(), metric.WithAttributes(moduleName, featureName, destinationVerb))

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
