package pubsub

import (
	"context"
	"time"

	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/backend/controller/scheduledtask"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/jpillora/backoff"
)

const (
	// Events can be added simultaneously, which can cause events with out of order create_at values
	// By adding a delay, we ensure that by the time we read the events, no new events will be added
	// with earlier created_at values.
	eventConsumptionDelay = 500 * time.Millisecond
)

type DAL interface {
	ProgressSubscriptions(ctx context.Context, eventConsumptionDelay time.Duration) (count int, err error)
	CompleteEventForSubscription(ctx context.Context, module, name string) error
}

type Scheduler interface {
	Singleton(retry backoff.Backoff, job scheduledtask.Job)
	Parallel(retry backoff.Backoff, job scheduledtask.Job)
}

type AsyncCallListener interface {
	AsyncCallWasAdded(ctx context.Context)
}

type Manager struct {
	dal               DAL
	scheduler         Scheduler
	asyncCallListener AsyncCallListener
}

func New(ctx context.Context, dal *dal.DAL, scheduler Scheduler, asyncCallListener AsyncCallListener) *Manager {
	m := &Manager{
		dal:               dal,
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

func (m *Manager) progressSubscriptions(ctx context.Context) (time.Duration, error) {
	count, err := m.dal.ProgressSubscriptions(ctx, eventConsumptionDelay)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		// notify controller that we added an async call
		go func() {
			m.asyncCallListener.AsyncCallWasAdded(ctx)
		}()
	}
	return time.Second, err
}

// OnCallCompletion is called within a transaction after an async call has completed to allow the subscription state to be updated.
func (m *Manager) OnCallCompletion(ctx context.Context, tx *dal.Tx, origin dal.AsyncOriginPubSub, failed bool) error {
	return m.dal.CompleteEventForSubscription(ctx, origin.Subscription.Module, origin.Subscription.Name)
}

// AsyncCallDidCommit is called after an subscription's async call has been completed and committed to the database.
func (m *Manager) AsyncCallDidCommit(ctx context.Context, origin dal.AsyncOriginPubSub) {
	if _, err := m.progressSubscriptions(ctx); err != nil {
		log.FromContext(ctx).Errorf(err, "failed to progress subscriptions")
	}
}
