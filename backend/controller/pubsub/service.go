package pubsub

import (
	"context"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/jpillora/backoff"

	"github.com/TBD54566975/ftl/backend/controller/async"
	"github.com/TBD54566975/ftl/backend/controller/encryption"
	"github.com/TBD54566975/ftl/backend/controller/pubsub/internal/dal"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/model"

	"github.com/TBD54566975/ftl/backend/controller/scheduledtask"
	"github.com/TBD54566975/ftl/internal/log"
)

const (
	// Events can be added simultaneously, which can cause events with out of order create_at values
	// By adding a delay, we ensure that by the time we read the events, no new events will be added
	// with earlier created_at values.
	eventConsumptionDelay = 200 * time.Millisecond
)

type Scheduler interface {
	Singleton(retry backoff.Backoff, job scheduledtask.Job)
	Parallel(retry backoff.Backoff, job scheduledtask.Job)
}

type AsyncCallListener interface {
	AsyncCallWasAdded(ctx context.Context)
}

type Service struct {
	dal               *dal.DAL
	scheduler         Scheduler
	asyncCallListener optional.Option[AsyncCallListener]
}

func New(conn libdal.Connection, encryption *encryption.Service, scheduler Scheduler, asyncCallListener optional.Option[AsyncCallListener]) *Service {
	m := &Service{
		dal:               dal.New(conn, encryption),
		scheduler:         scheduler,
		asyncCallListener: asyncCallListener,
	}
	m.scheduler.Parallel(backoff.Backoff{
		Min:    1 * time.Second,
		Max:    5 * time.Second,
		Jitter: true,
		Factor: 1.5,
	}, m.progressSubscriptions)
	return m
}

func (s *Service) progressSubscriptions(ctx context.Context) (time.Duration, error) {
	count, err := s.dal.ProgressSubscriptions(ctx, eventConsumptionDelay)
	if err != nil {
		return 0, fmt.Errorf("progress subscriptions: %w", err)
	}
	if count > 0 {
		// notify controller that we added an async call
		if listener, ok := s.asyncCallListener.Get(); ok {
			listener.AsyncCallWasAdded(ctx)
		}
	}
	return time.Second, nil
}

func (s *Service) PublishEventForTopic(ctx context.Context, module, topic, caller string, payload []byte) error {
	err := s.dal.PublishEventForTopic(ctx, module, topic, caller, payload)
	if err != nil {
		return fmt.Errorf("%s.%s: publish: %w", module, topic, err)
	}
	return nil
}

func (s *Service) ResetSubscription(ctx context.Context, module, name string) (err error) {
	err = s.dal.ResetSubscription(ctx, module, name)
	if err != nil {
		return fmt.Errorf("%s.%s: reset: %w", module, name, err)
	}
	return nil
}

// OnCallCompletion is called within a transaction after an async call has completed to allow the subscription state to be updated.
func (s *Service) OnCallCompletion(ctx context.Context, tx libdal.Connection, origin async.AsyncOriginPubSub, failed bool, isFinalResult bool) error {
	if !isFinalResult {
		// Wait for the async call's retries to complete before progressing the subscription
		return nil
	}
	err := s.dal.Adopt(tx).CompleteEventForSubscription(ctx, origin.Subscription.Module, origin.Subscription.Name)
	if err != nil {
		return fmt.Errorf("%s: complete: %w", origin, err)
	}
	return nil
}

// AsyncCallDidCommit is called after a subscription's async call has been completed and committed to the database.
func (s *Service) AsyncCallDidCommit(ctx context.Context, origin async.AsyncOriginPubSub) {
	if _, err := s.progressSubscriptions(ctx); err != nil {
		log.FromContext(ctx).Errorf(err, "failed to progress subscriptions")
	}
}

func (s *Service) CreateSubscriptions(ctx context.Context, key model.DeploymentKey, module *schema.Module) error {
	err := s.dal.CreateSubscriptions(ctx, key, module)
	if err != nil {
		return fmt.Errorf("create subscriptions: %w", err)
	}
	return nil
}

func (s *Service) RemoveSubscriptionsAndSubscribers(ctx context.Context, key model.DeploymentKey) error {
	err := s.dal.RemoveSubscriptionsAndSubscribers(ctx, key)
	if err != nil {
		return fmt.Errorf("remove subscriptions and subscribers: %w", err)
	}
	return nil
}

func (s *Service) CreateSubscribers(ctx context.Context, key model.DeploymentKey, module *schema.Module) error {
	err := s.dal.CreateSubscribers(ctx, key, module)
	if err != nil {
		return fmt.Errorf("create subscribers: %w", err)
	}
	return nil
}
