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
	pubsubMeterName                = "ftl.pubsub"
	pubsubTopicNameAttribute       = "ftl.pubsub.topic.name"
	pubsubSubscriptionRefAttribute = "ftl.pubsub.subscription.ref"
	pubsubSubscriberRefAttribute   = "ftl.pubsub.subscriber.sink.ref"
)

type PubSubMetrics struct {
	meter            metric.Meter
	published        metric.Int64Counter
	subscriberCalled metric.Int64Counter
}

func initPubSubMetrics() (*PubSubMetrics, error) {
	result := &PubSubMetrics{}
	var errs error
	var err error

	result.meter = otel.Meter(pubsubMeterName)

	counterName := fmt.Sprintf("%s.published", pubsubMeterName)
	if result.published, err = result.meter.Int64Counter(
		counterName,
		metric.WithDescription("the number of times that an event is published to a topic")); err != nil {
		result.published, errs = handleInitCounterError(errs, err, counterName)
	}

	counterName = fmt.Sprintf("%s.subscriber.called", pubsubMeterName)
	if result.subscriberCalled, err = result.meter.Int64Counter(
		counterName,
		metric.WithDescription("the number of times that a pubsub event has been enqueued to asynchronously send to a subscriber")); err != nil {
		result.subscriberCalled, errs = handleInitCounterError(errs, err, counterName)
	}

	return result, nil
}

func handleInitCounterError(errs error, err error, counterName string) (metric.Int64Counter, error) {
	return noop.Int64Counter{}, errors.Join(errs, fmt.Errorf("%q counter init failed; falling back to noop: %w", counterName, err))
}

func (m *PubSubMetrics) Published(ctx context.Context, module, topic string) {
	m.published.Add(ctx, 1, metric.WithAttributes(
		attribute.String(observability.ModuleNameAttribute, module),
		attribute.String(pubsubTopicNameAttribute, topic),
	))
}

func (m *PubSubMetrics) SubscriberCalled(ctx context.Context, topic string, subscription, sink schema.RefKey) {
	m.subscriberCalled.Add(ctx, 1, metric.WithAttributes(
		attribute.String(observability.ModuleNameAttribute, sink.Module),
		attribute.String(pubsubTopicNameAttribute, topic),
		attribute.String(pubsubSubscriptionRefAttribute, subscription.String()),
		attribute.String(pubsubSubscriberRefAttribute, sink.String()),
	))
}
