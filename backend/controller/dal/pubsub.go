package dal

import (
	"context"
	"fmt"
	"time"

	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/slices"
)

func (d *DAL) PublishEventForTopic(ctx context.Context, module, topic string, payload []byte) error {
	err := d.db.PublishEventForTopic(ctx, sql.PublishEventForTopicParams{
		Key:     model.NewTopicEventKey(module, topic),
		Module:  module,
		Topic:   topic,
		Payload: payload,
	})
	if err != nil {
		return translatePGError(err)
	}
	return nil
}

func (d *DAL) GetSubscriptionsNeedingUpdate(ctx context.Context) ([]model.Subscription, error) {
	rows, err := d.db.GetSubscriptionsNeedingUpdate(ctx)
	if err != nil {
		return nil, translatePGError(err)
	}
	return slices.Map(rows, func(row sql.GetSubscriptionsNeedingUpdateRow) model.Subscription {
		return model.Subscription{
			Name:   row.Name,
			Key:    row.Key,
			Topic:  row.Topic,
			Cursor: row.Cursor,
		}
	}), nil
}

func (d *DAL) ProgressSubscriptions(ctx context.Context, eventConsumptionDelay time.Duration) (count int, err error) {
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
		return 0, fmt.Errorf("could not get subscriptions to progress: %w", translatePGError(err))
	}

	successful := 0
	for _, subscription := range subs {
		nextCursor, err := tx.db.GetNextEventForSubscription(ctx, eventConsumptionDelay, subscription.Topic, subscription.Cursor)
		if err != nil {
			return 0, fmt.Errorf("failed to get next cursor: %w", translatePGError(err))
		}
		nextCursorKey, ok := nextCursor.Event.Get()
		if !ok {
			return 0, fmt.Errorf("could not find event to progress subscription: %w", translatePGError(err))
		}
		if !nextCursor.Ready {
			logger.Tracef("Skipping subscription %s because event is too new", subscription.Key)
			continue
		}

		subscriber, err := tx.db.GetRandomSubscriber(ctx, subscription.Key)
		if err != nil {
			logger.Tracef("no subscriber for subscription %s", subscription.Key)
			continue
		}

		err = tx.db.BeginConsumingTopicEvent(ctx, subscription.Key, nextCursorKey)
		if err != nil {
			return 0, fmt.Errorf("failed to progress subscription: %w", translatePGError(err))
		}

		origin := AsyncOriginPubSub{
			Subscription: schema.RefKey{
				Module: subscription.Key.Payload.Module,
				Name:   subscription.Key.Payload.Name,
			},
		}
		_, err = tx.db.CreateAsyncCall(ctx, sql.CreateAsyncCallParams{
			Verb:              subscriber.Sink,
			Origin:            origin.String(),
			Request:           nextCursor.Payload,
			RemainingAttempts: subscriber.RetryAttempts,
			Backoff:           subscriber.Backoff,
			MaxBackoff:        subscriber.MaxBackoff,
		})
		if err != nil {
			return 0, fmt.Errorf("failed to schedule async task for subscription: %w", translatePGError(err))
		}
		successful++
	}
	return successful, nil
}

func (d *DAL) CompleteEventForSubscription(ctx context.Context, module, name string) error {
	err := d.db.CompleteEventForSubscription(ctx, name, module)
	if err != nil {
		return fmt.Errorf("failed to complete event for subscription: %w", translatePGError(err))
	}
	return nil
}

func (d *DAL) createSubscriptions(ctx context.Context, tx *sql.Tx, key model.DeploymentKey, module *schema.Module) error {
	for _, decl := range module.Decls {
		s, ok := decl.(*schema.Subscription)
		if !ok {
			continue
		}
		if !hasSubscribers(s, module.Decls) {
			// Ignore subscriptions without subscribers
			// This ensures that controllers don't endlessly try to progress subscriptions without subscribers
			// https://github.com/TBD54566975/ftl/issues/1685
			//
			// It does mean that a subscription will reset to the topic's head if all subscribers are removed and then later re-added
			continue
		}
		if err := tx.UpsertSubscription(ctx, sql.UpsertSubscriptionParams{
			Key:         model.NewSubscriptionKey(module.Name, s.Name),
			Module:      module.Name,
			Deployment:  key,
			TopicModule: s.Topic.Module,
			TopicName:   s.Topic.Name,
			Name:        s.Name,
		}); err != nil {
			return fmt.Errorf("could not insert subscription: %w", translatePGError(err))
		}
	}
	return nil
}

func hasSubscribers(subscription *schema.Subscription, decls []schema.Decl) bool {
	for _, d := range decls {
		verb, ok := d.(*schema.Verb)
		if !ok {
			continue
		}
		for _, md := range verb.Metadata {
			subscriber, ok := md.(*schema.MetadataSubscriber)
			if !ok {
				continue
			}
			if subscriber.Name == subscription.Name {
				return true
			}
		}
	}
	return false
}

func (d *DAL) createSubscribers(ctx context.Context, tx *sql.Tx, key model.DeploymentKey, module *schema.Module) error {
	for _, decl := range module.Decls {
		v, ok := decl.(*schema.Verb)
		if !ok {
			continue
		}
		for _, md := range v.Metadata {
			s, ok := md.(*schema.MetadataSubscriber)
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
			err = tx.InsertSubscriber(ctx, sql.InsertSubscriberParams{
				Key:              model.NewSubscriberKey(module.Name, s.Name, v.Name),
				Module:           module.Name,
				SubscriptionName: s.Name,
				Deployment:       key,
				Sink:             sinkRef,
				RetryAttempts:    int32(retryParams.Count),
				Backoff:          retryParams.MinBackoff,
				MaxBackoff:       retryParams.MaxBackoff,
			})
			if err != nil {
				return fmt.Errorf("could not insert subscriber: %w", translatePGError(err))
			}
		}
	}
	return nil
}

func (d *DAL) removeSubscriptionsAndSubscribers(ctx context.Context, tx *sql.Tx, key model.DeploymentKey) error {
	if err := tx.DeleteSubscriptions(ctx, key); err != nil {
		return fmt.Errorf("could not delete old subscriptions: %w", translatePGError(err))
	}
	if err := tx.DeleteSubscribers(ctx, key); err != nil {
		return fmt.Errorf("could not delete old subscribers: %w", translatePGError(err))
	}
	return nil
}
