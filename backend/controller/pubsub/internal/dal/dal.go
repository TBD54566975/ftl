package dal

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/async"
	"github.com/TBD54566975/ftl/backend/controller/encryption"
	"github.com/TBD54566975/ftl/backend/controller/encryption/api"
	"github.com/TBD54566975/ftl/backend/controller/observability"
	dalsql "github.com/TBD54566975/ftl/backend/controller/pubsub/internal/sql"
	"github.com/TBD54566975/ftl/backend/controller/sql/sqltypes"
	"github.com/TBD54566975/ftl/backend/controller/timeline"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/slices"
)

type DAL struct {
	*libdal.Handle[DAL]
	db         dalsql.Querier
	encryption *encryption.Service
}

func New(conn libdal.Connection, encryption *encryption.Service) *DAL {
	return &DAL{
		Handle: libdal.New(conn, func(h *libdal.Handle[DAL]) *DAL {
			return &DAL{
				Handle:     h,
				db:         dalsql.New(h.Connection),
				encryption: encryption,
			}
		}),
		db:         dalsql.New(conn),
		encryption: encryption,
	}
}

func (d *DAL) PublishEventForTopic(ctx context.Context, module, topic, caller string, payload []byte) error {
	var encryptedPayload api.EncryptedAsyncColumn
	err := d.encryption.Encrypt(payload, &encryptedPayload)
	if err != nil {
		return fmt.Errorf("failed to encrypt payload: %w", err)
	}

	// Store the current otel context with the event
	jsonOc, err := observability.RetrieveTraceContextFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve trace context: %w", err)
	}

	// Store the request key that initiated this publish, this will eventually
	// become the parent request key of the subscriber call
	requestKey := ""
	if rk, err := rpc.RequestKeyFromContext(ctx); err == nil {
		if rk, ok := rk.Get(); ok {
			requestKey = rk.String()
		}
	} else {
		return fmt.Errorf("failed to get request key: %w", err)
	}

	err = d.db.PublishEventForTopic(ctx, dalsql.PublishEventForTopicParams{
		Key:          model.NewTopicEventKey(module, topic),
		Module:       module,
		Topic:        topic,
		Caller:       caller,
		Payload:      encryptedPayload,
		RequestKey:   requestKey,
		TraceContext: jsonOc,
	})
	observability.PubSub.Published(ctx, module, topic, caller, err)
	if err != nil {
		return libdal.TranslatePGError(err)
	}
	return nil
}

func (d *DAL) GetSubscriptionsNeedingUpdate(ctx context.Context) ([]model.Subscription, error) {
	rows, err := d.db.GetSubscriptionsNeedingUpdate(ctx)
	if err != nil {
		return nil, libdal.TranslatePGError(err)
	}
	return slices.Map(rows, func(row dalsql.GetSubscriptionsNeedingUpdateRow) model.Subscription {
		return model.Subscription{
			Name:   row.Name,
			Key:    row.Key,
			Topic:  row.Topic,
			Cursor: row.Cursor,
		}
	}), nil
}

func (d *DAL) ProgressSubscriptions(ctx context.Context, eventConsumptionDelay time.Duration, timelineSvc *timeline.Service) (count int, err error) {
	tx, err := d.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	logger := log.FromContext(ctx)

	// get subscriptions needing update
	// also gets a lock on the subscription, and skips any subscriptions locked by others
	subs, err := tx.db.GetSubscriptionsNeedingUpdate(ctx)
	if err != nil {
		return 0, fmt.Errorf("could not get subscriptions to progress: %w", libdal.TranslatePGError(err))
	}

	successful := 0
	for _, subscription := range subs {
		now := time.Now().UTC()
		enqueueTimelineEvent := func(destVerb optional.Option[schema.RefKey], err optional.Option[string]) {
			timelineSvc.EnqueueEvent(ctx, &timeline.PubSubConsume{
				DeploymentKey: subscription.DeploymentKey,
				RequestKey:    subscription.RequestKey,
				Time:          now,
				DestVerb:      destVerb,
				Topic:         subscription.Topic.Payload.Name,
				Error:         err,
			})
		}

		nextCursor, err := tx.db.GetNextEventForSubscription(ctx, sqltypes.Duration(eventConsumptionDelay), subscription.Topic, subscription.Cursor)
		if err != nil {
			observability.PubSub.PropagationFailed(ctx, "GetNextEventForSubscription", subscription.Topic.Payload, nextCursor.Caller, subscriptionRef(subscription), optional.None[schema.RefKey]())
			err = fmt.Errorf("failed to get next cursor: %w", libdal.TranslatePGError(err))
			enqueueTimelineEvent(optional.None[schema.RefKey](), optional.Some(err.Error()))
			return 0, err
		}
		payload, ok := nextCursor.Payload.Get()
		if !ok {
			observability.PubSub.PropagationFailed(ctx, "GetNextEventForSubscription-->Payload.Get", subscription.Topic.Payload, nextCursor.Caller, subscriptionRef(subscription), optional.None[schema.RefKey]())
			err = fmt.Errorf("could not find payload to progress subscription: %w", libdal.TranslatePGError(err))
			enqueueTimelineEvent(optional.None[schema.RefKey](), optional.Some(err.Error()))
			return 0, err
		}
		nextCursorKey, ok := nextCursor.Event.Get()
		if !ok {
			observability.PubSub.PropagationFailed(ctx, "GetNextEventForSubscription-->Event.Get", subscription.Topic.Payload, nextCursor.Caller, subscriptionRef(subscription), optional.None[schema.RefKey]())
			err = fmt.Errorf("could not find event to progress subscription: %w", libdal.TranslatePGError(err))
			enqueueTimelineEvent(optional.None[schema.RefKey](), optional.Some(err.Error()))
			return 0, err
		}
		if !nextCursor.Ready {
			logger.Tracef("Skipping subscription %s because event is too new", subscription.Key)
			enqueueTimelineEvent(optional.None[schema.RefKey](), optional.Some(fmt.Sprintf("Skipping subscription %s because event is too new", subscription.Key)))
			continue
		}

		subscriber, err := tx.db.GetRandomSubscriber(ctx, subscription.Key)
		if err != nil {
			logger.Tracef("no subscriber for subscription %s", subscription.Key)
			enqueueTimelineEvent(optional.None[schema.RefKey](), optional.Some(fmt.Sprintf("no subscriber for subscription %s", subscription.Key)))
			continue
		}

		err = tx.db.BeginConsumingTopicEvent(ctx, subscription.Key, nextCursorKey)
		if err != nil {
			observability.PubSub.PropagationFailed(ctx, "BeginConsumingTopicEvent", subscription.Topic.Payload, nextCursor.Caller, subscriptionRef(subscription), optional.Some(subscriber.Sink))
			err = fmt.Errorf("failed to progress subscription: %w", libdal.TranslatePGError(err))
			enqueueTimelineEvent(optional.Some(subscriber.Sink), optional.Some(err.Error()))
			return 0, err
		}

		origin := async.AsyncOriginPubSub{
			Subscription: schema.RefKey{
				Module: subscription.Key.Payload.Module,
				Name:   subscription.Key.Payload.Name,
			},
		}

		_, err = tx.db.CreateAsyncCall(ctx, dalsql.CreateAsyncCallParams{
			ScheduledAt:       time.Now(),
			Verb:              subscriber.Sink,
			Origin:            origin.String(),
			Request:           payload, // already encrypted
			RemainingAttempts: subscriber.RetryAttempts,
			Backoff:           subscriber.Backoff,
			MaxBackoff:        subscriber.MaxBackoff,
			ParentRequestKey:  nextCursor.RequestKey,
			TraceContext:      nextCursor.TraceContext.RawMessage,
			CatchVerb:         subscriber.CatchVerb,
		})
		observability.AsyncCalls.Created(ctx, subscriber.Sink, subscriber.CatchVerb, origin.String(), int64(subscriber.RetryAttempts), err)
		if err != nil {
			observability.PubSub.PropagationFailed(ctx, "CreateAsyncCall", subscription.Topic.Payload, nextCursor.Caller, subscriptionRef(subscription), optional.Some(subscriber.Sink))
			err = fmt.Errorf("failed to schedule async task for subscription: %w", libdal.TranslatePGError(err))
			enqueueTimelineEvent(optional.Some(subscriber.Sink), optional.Some(err.Error()))
			return 0, err
		}

		observability.PubSub.SinkCalled(ctx, subscription.Topic.Payload, nextCursor.Caller, subscriptionRef(subscription), subscriber.Sink)
		enqueueTimelineEvent(optional.Some(subscriber.Sink), optional.None[string]())
		successful++
	}

	if successful > 0 {
		// If no async calls were successfully created, then there is no need to
		// potentially increment the queue depth gauge.
		queueDepth, err := tx.db.AsyncCallQueueDepth(ctx)
		if err == nil {
			// Don't error out of progressing subscriptions just over a queue depth
			// retrieval error because this is only used for an observability gauge.
			observability.AsyncCalls.RecordQueueDepth(ctx, queueDepth)
		}
	}

	return successful, nil
}

func subscriptionRef(subscription dalsql.GetSubscriptionsNeedingUpdateRow) schema.RefKey {
	return schema.RefKey{Module: subscription.Key.Payload.Module, Name: subscription.Name}
}

func (d *DAL) CompleteEventForSubscription(ctx context.Context, module, name string) error {
	err := d.db.CompleteEventForSubscription(ctx, name, module)
	if err != nil {
		return fmt.Errorf("failed to complete event for subscription: %w", libdal.TranslatePGError(err))
	}
	return nil
}

// ResetSubscription resets the subscription cursor to the topic's head.
func (d *DAL) ResetSubscription(ctx context.Context, module, name string) (err error) {
	tx, err := d.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	subscription, err := tx.db.GetSubscription(ctx, name, module)
	if err != nil {
		if libdal.IsNotFound(err) {
			return fmt.Errorf("subscription %s.%s not found", module, name)
		}
		return fmt.Errorf("could not fetch subscription: %w", libdal.TranslatePGError(err))
	}

	topic, err := tx.db.GetTopic(ctx, subscription.TopicID)
	if err != nil {
		return fmt.Errorf("could not fetch topic: %w", libdal.TranslatePGError(err))
	}

	headEventID, ok := topic.Head.Get()
	if !ok {
		return fmt.Errorf("no events published to topic %s", topic.Name)
	}

	headEvent, err := tx.db.GetTopicEvent(ctx, headEventID)
	if err != nil {
		return fmt.Errorf("could not fetch topic head: %w", libdal.TranslatePGError(err))
	}

	err = tx.db.SetSubscriptionCursor(ctx, subscription.Key, headEvent.Key)
	if err != nil {
		return fmt.Errorf("failed to reset subscription: %w", libdal.TranslatePGError(err))
	}

	return nil
}

func (d *DAL) CreateSubscriptions(ctx context.Context, key model.DeploymentKey, module *schema.Module) error {
	logger := log.FromContext(ctx)

	for verb := range slices.FilterVariants[*schema.Verb](module.Decls) {
		subscriber, ok := slices.FindVariant[*schema.MetadataSubscriber](verb.Metadata)
		if !ok {
			continue
		}
		subscriptionKey := model.NewSubscriptionKey(module.Name, verb.Name)
		result, err := d.db.UpsertSubscription(ctx, dalsql.UpsertSubscriptionParams{
			Key:         subscriptionKey,
			Module:      module.Name,
			Deployment:  key,
			TopicModule: subscriber.Topic.Module,
			TopicName:   subscriber.Topic.Name,
			Name:        verb.Name,
		})
		if err != nil {
			return fmt.Errorf("could not insert subscription: %w", libdal.TranslatePGError(err))
		}
		if result.Inserted {
			logger.Debugf("Inserted subscription %s for %s", subscriptionKey, key)
		} else {
			logger.Debugf("Updated subscription %s to %s", subscriptionKey, key)
		}
	}
	return nil
}

func (d *DAL) CreateSubscribers(ctx context.Context, key model.DeploymentKey, module *schema.Module) error {
	logger := log.FromContext(ctx)
	for _, decl := range module.Decls {
		v, ok := decl.(*schema.Verb)
		if !ok {
			continue
		}
		for _, md := range v.Metadata {
			_, ok := md.(*schema.MetadataSubscriber)
			if !ok {
				continue
			}
			sinkRef := schema.RefKey{
				Module: module.Name,
				Name:   v.Name,
			}
			retryParams := schema.RetryParams{}
			var err error
			if retryMd, ok := slices.FindVariant[*schema.MetadataRetry](v.Metadata); ok {
				retryParams, err = retryMd.RetryParams()
				if err != nil {
					return fmt.Errorf("could not parse retry parameters for %q: %w", v.Name, err)
				}
			}
			subscriberKey := model.NewSubscriberKey(module.Name, v.Name, v.Name)
			err = d.db.InsertSubscriber(ctx, dalsql.InsertSubscriberParams{
				Key:              subscriberKey,
				Module:           module.Name,
				SubscriptionName: v.Name,
				Deployment:       key,
				Sink:             sinkRef,
				RetryAttempts:    int32(retryParams.Count),
				Backoff:          sqltypes.Duration(retryParams.MinBackoff),
				MaxBackoff:       sqltypes.Duration(retryParams.MaxBackoff),
				CatchVerb:        retryParams.Catch,
			})
			if err != nil {
				return fmt.Errorf("could not insert subscriber: %w", libdal.TranslatePGError(err))
			}
			logger.Debugf("Inserted subscriber %s for %s", subscriberKey, key)
		}
	}
	return nil
}

func (d *DAL) RemoveSubscriptionsAndSubscribers(ctx context.Context, key model.DeploymentKey) error {
	logger := log.FromContext(ctx)

	subscribers, err := d.db.DeleteSubscribers(ctx, key)
	if err != nil {
		return fmt.Errorf("could not delete old subscribers: %w", libdal.TranslatePGError(err))
	}
	if len(subscribers) > 0 {
		logger.Debugf("Deleted subscribers for %s: %s", key, strings.Join(slices.Map(subscribers, func(key model.SubscriberKey) string { return key.String() }), ", "))
	}

	subscriptions, err := d.db.DeleteSubscriptions(ctx, key)
	if err != nil {
		return fmt.Errorf("could not delete old subscriptions: %w", libdal.TranslatePGError(err))
	}
	if len(subscriptions) > 0 {
		logger.Debugf("Deleted subscriptions for %s: %s", key, strings.Join(slices.Map(subscriptions, func(key model.SubscriptionKey) string { return key.String() }), ", "))
	}

	return nil
}
