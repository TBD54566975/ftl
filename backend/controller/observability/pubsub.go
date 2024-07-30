package observability

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/internal/observability"
)

const (
	pubsubMeterName              = "ftl.pubsub"
	pubsubTopicNameAttr          = "ftl.pubsub.topic.name"
	pubsubSubscriptionRefAttr    = "ftl.pubsub.subscription.ref"
	pubsubSubscriptionModuleAttr = "ftl.pubsub.subscription.module.name"
	pubsubSubscriberRefAttr      = "ftl.pubsub.subscriber.sink.ref"
	pubsubSubscriberModuleAttr   = "ftl.pubsub.subscriber.sink.module.name"
	pubsubFailedOperationAttr    = "ftl.pubsub.propagation.failed_operation"
)

type PubSubMetrics struct {
	meter             metric.Meter
	published         metric.Int64Counter
	propagationFailed metric.Int64Counter
	subscriberCalled  metric.Int64Counter
}

func initPubSubMetrics() (*PubSubMetrics, error) {
	result := &PubSubMetrics{}
	var errs error
	var err error

	result.meter = otel.Meter(pubsubMeterName)

	counterName := fmt.Sprintf("%s.published", pubsubMeterName)
	if result.published, err = result.meter.Int64Counter(
		counterName,
		metric.WithUnit("1"),
		metric.WithDescription("the number of times that an event is published to a topic")); err != nil {
		errs = handleInitCounterError(errs, err, counterName)
		result.published = noop.Int64Counter{}
	}

	counterName = fmt.Sprintf("%s.propagation.failed", pubsubMeterName)
	if result.propagationFailed, err = result.meter.Int64Counter(
		counterName,
		metric.WithUnit("1"),
		metric.WithDescription("the number of times that subscriptions fail to progress")); err != nil {
		errs = handleInitCounterError(errs, err, counterName)
		result.propagationFailed = noop.Int64Counter{}
	}

	counterName = fmt.Sprintf("%s.subscriber.called", pubsubMeterName)
	if result.subscriberCalled, err = result.meter.Int64Counter(
		counterName,
		metric.WithUnit("1"),
		metric.WithDescription("the number of times that a pubsub event has been enqueued to asynchronously send to a subscriber")); err != nil {
		errs = handleInitCounterError(errs, err, counterName)
		result.subscriberCalled = noop.Int64Counter{}
	}

	return result, errs
}

func handleInitCounterError(errs error, err error, counterName string) error {
	return errors.Join(errs, fmt.Errorf("%q counter init failed; falling back to noop: %w", counterName, err))
}

func (m *PubSubMetrics) Published(ctx context.Context, module, topic string, succeeded bool) {
	m.published.Add(ctx, 1, metric.WithAttributes(
		attribute.String(observability.ModuleNameAttribute, module),
		attribute.String(pubsubTopicNameAttr, topic),
		attribute.Bool(observability.StatusSucceededAttribute, succeeded),
	))
}

func (m *PubSubMetrics) PropagationFailed(ctx context.Context, failedOp, topic string, subscription schema.RefKey, sink ftl.Option[schema.RefKey]) {
	attrs := []attribute.KeyValue{
		attribute.String(pubsubFailedOperationAttr, failedOp),
		attribute.String(pubsubTopicNameAttr, topic),
		attribute.String(pubsubSubscriptionRefAttr, subscription.String()),
		attribute.String(pubsubSubscriptionModuleAttr, subscription.Module),
	}

	if sinkRef, ok := sink.Get(); ok {
		attrs = append(attrs, attribute.String(pubsubSubscriberRefAttr, sinkRef.String()))
		attrs = append(attrs, attribute.String(pubsubSubscriberModuleAttr, sinkRef.Module))
	}

	m.propagationFailed.Add(ctx, 1, metric.WithAttributes(attrs...))
}

func (m *PubSubMetrics) SubscriberCalled(ctx context.Context, topic string, subscription, sink schema.RefKey) {
	m.subscriberCalled.Add(ctx, 1, metric.WithAttributes(
		attribute.String(pubsubTopicNameAttr, topic),
		attribute.String(pubsubSubscriptionRefAttr, subscription.String()),
		attribute.String(pubsubSubscriptionModuleAttr, subscription.Module),
		attribute.String(pubsubSubscriberRefAttr, sink.String()),
		attribute.String(pubsubSubscriberModuleAttr, sink.Module),
	))
}
