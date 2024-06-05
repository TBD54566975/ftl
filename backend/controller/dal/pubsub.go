package dal

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/controller/sql"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/alecthomas/types/optional"
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

func (d *DAL) ProgressSubscription(ctx context.Context, subscription model.Subscription) error {
	tx, err := d.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.CommitOrRollback(ctx, &err)

	nextCursor, err := tx.db.GetNextEventForSubscription(ctx, subscription.Topic, subscription.Cursor)
	if err != nil {
		return fmt.Errorf("failed to get next cursor: %w", translatePGError(err))
	}

	result, err := tx.db.LockSubscriptionAndGetSink(ctx, subscription.Key, subscription.Cursor)
	if err != nil {
		return fmt.Errorf("failed to get lock on subscription: %w", translatePGError(err))
	}

	err = tx.db.BeginConsumingTopicEvent(ctx, optional.Some(result.SubscriptionID), nextCursor.Event)
	if err != nil {
		return fmt.Errorf("failed to progress subscription: %w", translatePGError(err))
	}

	origin := AsyncOriginPubSub{
		Subscription: schema.RefKey{
			Module: subscription.Key.Payload.Module,
			Name:   subscription.Key.Payload.Name,
		},
	}
	_, err = tx.db.CreateAsyncCall(ctx, sql.CreateAsyncCallParams{
		Verb:    result.Sink,
		Origin:  origin.String(),
		Request: nextCursor.Payload,
	})
	if err != nil {
		return fmt.Errorf("failed to schedule async task for subscription: %w", translatePGError(err))
	}
	return nil
}

func (d *DAL) CompleteEventForSubscription(ctx context.Context, module, name string) error {
	err := d.db.CompleteEventForSubscription(ctx, name, module)
	if err != nil {
		return fmt.Errorf("failed to complete event for subscription: %w", translatePGError(err))
	}
	return nil
}
