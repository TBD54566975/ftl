package pubsub

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/jpillora/backoff"

	"github.com/TBD54566975/ftl/backend/controller/async"
	"github.com/TBD54566975/ftl/backend/controller/pubsub/internal/dal"
	"github.com/TBD54566975/ftl/backend/controller/scheduledtask"
	"github.com/TBD54566975/ftl/backend/libdal"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/schema"
)

const (
	// Events can be added simultaneously, which can cause events with out of order create_at values
	// By adding a delay, we ensure that by the time we read the events, no new events will be added
	// with earlier created_at values.
	eventConsumptionDelay = 100 * time.Millisecond
)

type Scheduler interface {
	Singleton(name string, retry backoff.Backoff, job scheduledtask.Job)
	Parallel(name string, retry backoff.Backoff, job scheduledtask.Job)
}

type AsyncCallListener interface {
	AsyncCallWasAdded(ctx context.Context)
}

type Service struct {
	dal               *dal.DAL
	asyncCallListener optional.Option[AsyncCallListener]
	eventPublished    chan struct{}
}

func New(ctx context.Context, conn libdal.Connection, asyncCallListener optional.Option[AsyncCallListener]) *Service {
	m := &Service{
		dal:               dal.New(conn),
		asyncCallListener: asyncCallListener,
		eventPublished:    make(chan struct{}),
	}
	go m.poll(ctx)
	return m
}

// poll waits for an event to be published (incl eventConsumptionDelay) or for a poll interval to pass
func (s *Service) poll(ctx context.Context) {
	logger := log.FromContext(ctx).Scope("pubsub")
	var publishedAt optional.Option[time.Time]
	for {
		var publishTrigger <-chan time.Time
		if pub, ok := publishedAt.Get(); ok {
			publishTrigger = time.After(time.Until(pub.Add(eventConsumptionDelay)))
		}

		// poll interval with jitter (1s - 1.1s)
		poll := time.Millisecond * (time.Duration(rand.Float64())*(100.0) + 1000.0) //nolint:gosec

		select {
		case <-ctx.Done():
			return

		case <-s.eventPublished:
			// published an event, so now we wait for eventConsumptionDelay before trying to progress subscriptions
			if !publishedAt.Ok() {
				publishedAt = optional.Some(time.Now())
			}

		case <-publishTrigger:
			// an event has been published and we have waited for eventConsumptionDelay
			if err := s.progressSubscriptions(ctx); err != nil {
				logger.Warnf("%s", err)
			}
			publishedAt = optional.None[time.Time]()

		case <-time.After(poll):
			if err := s.progressSubscriptions(ctx); err != nil {
				logger.Warnf("%s", err)
			}
		}
	}
}

func (s *Service) progressSubscriptions(ctx context.Context) error {
	count, err := s.dal.ProgressSubscriptions(ctx, eventConsumptionDelay)
	if err != nil {
		return fmt.Errorf("progress subscriptions: %w", err)
	}
	if count > 0 {
		// notify controller that we added an async call
		if listener, ok := s.asyncCallListener.Get(); ok {
			listener.AsyncCallWasAdded(ctx)
		}
	}
	return nil
}

func (s *Service) PublishEventForTopic(ctx context.Context, module, topic, caller string, payload []byte) error {
	err := s.dal.PublishEventForTopic(ctx, module, topic, caller, payload)
	if err != nil {
		return fmt.Errorf("%s.%s: publish: %w", module, topic, err)
	}
	s.eventPublished <- struct{}{}
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
	if err := s.progressSubscriptions(ctx); err != nil {
		log.FromContext(ctx).Scope("pubsub").Errorf(err, "failed to progress subscriptions")
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
