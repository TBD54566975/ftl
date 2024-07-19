package controller

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

const name = "ftl.xyz/ftl/runner"

var (
	meter          = otel.Meter(name)
	requestsMetric metric.Int64Counter
	failureMetric  metric.Int64Counter
	latencyMetric  metric.Int64Histogram

	moduleNameAttr = func(name string) attribute.KeyValue {
		return attribute.String("ftl.module.name", name)
	}
	sourceVerbAttr = func(name string) attribute.KeyValue {
		return attribute.String("ftl.verb.src", name)
	}
	destinationVerbAttr = func(name string) attribute.KeyValue {
		return attribute.String("ftl.verb.dest", name)
	}
)

type Call struct {
	deploymentKey model.DeploymentKey
	requestKey    model.RequestKey
	startTime     time.Time
	destVerb      *schema.Ref
	callers       []*schema.Ref
	request       *ftlv1.CallRequest
	response      optional.Option[*ftlv1.CallResponse]
	callError     optional.Option[error]
}

func init() {
	requestsMetric, _ = meter.Int64Counter("ftl.call.requests",
		metric.WithDescription("number of verb calls"),
		metric.WithUnit("{count}"))

	failureMetric, _ = meter.Int64Counter("ftl.call.failures",
		metric.WithDescription("number of verb call failures"),
		metric.WithUnit("{count}"))

	latencyMetric, _ = meter.Int64Histogram("ftl.call.latency",
		metric.WithDescription("verb call latency"),
		metric.WithUnit("{ms}"))
}

func (s *Service) recordCall(ctx context.Context, call *Call) {
	logger := log.FromContext(ctx)
	var sourceVerb optional.Option[schema.Ref]
	var srcAttr attribute.KeyValue
	var moduleAttr attribute.KeyValue
	if len(call.callers) > 0 {
		sourceVerb = optional.Some(*call.callers[0])
		srcAttr = sourceVerbAttr(call.callers[0].Name)
		moduleAttr = moduleNameAttr(call.callers[0].Module)
	} else {
		srcAttr = sourceVerbAttr("unknown")
		moduleAttr = moduleNameAttr("unknown")
	}

	destAttr := destinationVerbAttr(call.destVerb.Name)

	requestsMetric.Add(ctx, 1, metric.WithAttributes(moduleAttr, srcAttr, destAttr))

	var errorStr optional.Option[string]
	var stack optional.Option[string]
	var responseBody []byte

	if callError, ok := call.callError.Get(); ok {
		errorStr = optional.Some(callError.Error())
		failureMetric.Add(ctx, 1, metric.WithAttributes(moduleAttr, srcAttr, destAttr))
	} else if response, ok := call.response.Get(); ok {
		responseBody = response.GetBody()
		if callError := response.GetError(); callError != nil {
			errorStr = optional.Some(callError.Message)
			stack = optional.Ptr(callError.Stack)
			failureMetric.Add(ctx, 1, metric.WithAttributes(moduleAttr, srcAttr, destAttr))
		}
	}

	latencyMetric.Record(ctx, time.Now().Sub(call.startTime).Milliseconds(), metric.WithAttributes(moduleAttr, srcAttr, destAttr))

	err := s.dal.InsertCallEvent(ctx, &dal.CallEvent{
		Time:          call.startTime,
		DeploymentKey: call.deploymentKey,
		RequestKey:    optional.Some(call.requestKey),
		Duration:      time.Since(call.startTime),
		SourceVerb:    sourceVerb,
		DestVerb:      *call.destVerb,
		Request:       call.request.GetBody(),
		Response:      responseBody,
		Error:         errorStr,
		Stack:         stack,
	})
	if err != nil {
		logger.Errorf(err, "failed to record call")
	}
}
