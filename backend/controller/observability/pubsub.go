package observability

import (
	"context"
	"fmt"

	"github.com/alecthomas/types/optional"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/observability"
)

// To learn more about how sinks and subscriptions work together, check out the
// https://tbd54566975.github.io/ftl/docs/reference/pubsub/
const (
	pubsubMeterName              = "ftl.pubsub"
	pubsubTopicRefAttr           = "ftl.pubsub.topic.ref"
	pubsubTopicModuleAttr        = "ftl.pubsub.topic.module.name"
	pubsubCallerVerbRefAttr      = "ftl.pubsub.publish.caller.verb.ref"
	pubsubSubscriptionRefAttr    = "ftl.pubsub.subscription.ref"
	pubsubSubscriptionModuleAttr = "ftl.pubsub.subscription.module.name"
	pubsubSinkRefAttr            = "ftl.pubsub.sink.ref"
	pubsubSinkModuleAttr         = "ftl.pubsub.sink.module.name"
	pubsubFailedOperationAttr    = "ftl.pubsub.propagation.failed_operation"
)

type PubSubMetrics struct {
	published         metric.Int64Counter
	propagationFailed metric.Int64Counter
	sinkCalled        metric.Int64Counter
}

func initPubSubMetrics() (*PubSubMetrics, error) {
	result := &PubSubMetrics{
		published:         noop.Int64Counter{},
		propagationFailed: noop.Int64Counter{},
		sinkCalled:        noop.Int64Counter{},
	}

	var err error
	meter := otel.Meter(pubsubMeterName)

	counterName := fmt.Sprintf("%s.published", pubsubMeterName)
	if result.published, err = meter.Int64Counter(
		counterName,
		metric.WithUnit("1"),
		metric.WithDescription("the number of times that an event is published to a topic")); err != nil {
		return nil, wrapErr(counterName, err)
	}

	counterName = fmt.Sprintf("%s.propagation.failed", pubsubMeterName)
	if result.propagationFailed, err = meter.Int64Counter(
		counterName,
		metric.WithUnit("1"),
		metric.WithDescription("the number of times that subscriptions fail to progress")); err != nil {
		return nil, wrapErr(counterName, err)
	}

	counterName = fmt.Sprintf("%s.sink.called", pubsubMeterName)
	if result.sinkCalled, err = meter.Int64Counter(
		counterName,
		metric.WithUnit("1"),
		metric.WithDescription("the number of times that a pubsub event has been enqueued to asynchronously send to a subscriber")); err != nil {
		return nil, wrapErr(counterName, err)
	}

	return result, nil
}

func (m *PubSubMetrics) Published(ctx context.Context, module, topic, caller string, maybeErr error) {
	attrs := []attribute.KeyValue{
		attribute.String(observability.ModuleNameAttribute, module),
		attribute.String(pubsubTopicRefAttr, schema.RefKey{Module: module, Name: topic}.String()),
		attribute.String(pubsubCallerVerbRefAttr, schema.RefKey{Module: module, Name: caller}.String()),
		observability.SuccessOrFailureStatusAttr(maybeErr == nil),
	}

	m.published.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (m *PubSubMetrics) PropagationFailed(ctx context.Context, failedOp string, topic model.TopicPayload, optCaller optional.Option[string], subscription schema.RefKey, sink optional.Option[schema.RefKey]) {
	attrs := []attribute.KeyValue{
		attribute.String(pubsubFailedOperationAttr, failedOp),
		attribute.String(pubsubTopicRefAttr, schema.RefKey{Module: topic.Module, Name: topic.Name}.String()),
		attribute.String(pubsubTopicModuleAttr, topic.Module),
		attribute.String(pubsubSubscriptionRefAttr, subscription.String()),
		attribute.String(pubsubSubscriptionModuleAttr, subscription.Module),
	}

	caller, ok := optCaller.Get()
	if ok {
		attrs = append(attrs, attribute.String(pubsubCallerVerbRefAttr, schema.RefKey{Module: topic.Module, Name: caller}.String()))
	}

	if sinkRef, ok := sink.Get(); ok {
		attrs = append(attrs, attribute.String(pubsubSinkRefAttr, sinkRef.String()))
		attrs = append(attrs, attribute.String(pubsubSinkModuleAttr, sinkRef.Module))
	}

	m.propagationFailed.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (m *PubSubMetrics) SinkCalled(ctx context.Context, topic model.TopicPayload, optCaller optional.Option[string], subscription, sink schema.RefKey) {
	attrs := []attribute.KeyValue{
		attribute.String(pubsubTopicRefAttr, schema.RefKey{Module: topic.Module, Name: topic.Name}.String()),
		attribute.String(pubsubTopicModuleAttr, topic.Module),
		attribute.String(pubsubSubscriptionRefAttr, subscription.String()),
		attribute.String(pubsubSubscriptionModuleAttr, subscription.Module),
		attribute.String(pubsubSinkRefAttr, sink.String()),
		attribute.String(pubsubSinkModuleAttr, sink.Module),
	}

	caller, ok := optCaller.Get()
	if ok {
		attrs = append(attrs, attribute.String(pubsubCallerVerbRefAttr, schema.RefKey{Module: topic.Module, Name: caller}.String()))
	}

	m.sinkCalled.Add(ctx, 1, metric.WithAttributes(attrs...))
}
