package pubsub

import (
	"context"
	"fmt"
	"time"

	"github.com/TBD54566975/ftl/backend/controller/dal"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/alecthomas/atomic"
	"github.com/serialx/hashring"
)

const (
	controllersPerSubscription = 2
)

type DAL interface {
	GetSubscriptionsNeedingUpdate(ctx context.Context) ([]model.Subscription, error)
	ProgressSubscription(ctx context.Context, subscription model.Subscription) error
	CompleteEventForSubscription(ctx context.Context, module, name string) error
}

type Manager struct {
	key           model.ControllerKey
	dal           DAL
	hashRingState atomic.Value[*hashRingState]
}

type hashRingState struct {
	hashRing    *hashring.HashRing
	controllers []dal.Controller
	idx         int
}

func New(ctx context.Context, key model.ControllerKey, dal *dal.DAL) *Manager {
	m := &Manager{
		key: key,
		dal: dal,
	}

	go m.watchForUpdates(ctx)
	return m
}

func (m *Manager) HandleTopicNotification() {

}

func (m *Manager) HandleEventNotification() {

}

// UpdatedControllerList synchronises the hash ring with the active controllers.
func (m *Manager) UpdatedControllerList(ctx context.Context, controllers []dal.Controller) {
	logger := log.FromContext(ctx).Scope("cron")
	controllerIdx := -1
	for idx, controller := range controllers {
		if controller.Key.String() == m.key.String() {
			controllerIdx = idx
			break
		}
	}
	if controllerIdx == -1 {
		logger.Tracef("controller %q not found in list of controllers", m.key)
	}

	oldState := m.hashRingState.Load()
	if oldState != nil && len(oldState.controllers) == len(controllers) {
		hasChanged := false
		for idx, new := range controllers {
			old := oldState.controllers[idx]
			if new.Key.String() != old.Key.String() {
				hasChanged = true
				break
			}
		}
		if !hasChanged {
			return
		}
	}

	hashRing := hashring.New(slices.Map(controllers, func(c dal.Controller) string { return c.Key.String() }))
	m.hashRingState.Store(&hashRingState{
		hashRing:    hashRing,
		controllers: controllers,
		idx:         controllerIdx,
	})
}

// isResponsibleForSubscription indicates whether a this service should be responsible for attempting jobs,
// or if enough other controllers will handle it. This allows us to spread the job load across controllers.
func (m *Manager) isResponsibleForSubscription(subscription model.Subscription) bool {
	hashringState := m.hashRingState.Load()
	if hashringState == nil {
		return true
	}

	initialKey, ok := hashringState.hashRing.GetNode(subscription.Key.String())
	if !ok {
		return true
	}

	initialIdx := -1
	for idx, controller := range hashringState.controllers {
		if controller.Key.String() == initialKey {
			initialIdx = idx
			break
		}
	}
	if initialIdx == -1 {
		return true
	}

	if initialIdx+controllersPerSubscription > len(hashringState.controllers) {
		// wraps around
		return hashringState.idx >= initialIdx || hashringState.idx < (initialIdx+controllersPerSubscription)-len(hashringState.controllers)
	}
	return hashringState.idx >= initialIdx && hashringState.idx < initialIdx+controllersPerSubscription
}

func (m *Manager) watchForUpdates(ctx context.Context) {
	logger := log.FromContext(ctx).Scope("pubsub")

	// TODO: handle events here. Currently a demo implementation
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second * 3):
			if err := m.progressSubscriptions(ctx); err != nil {
				logger.Errorf(err, "failed to progress subscriptions")
				continue
			}
		}
	}
}

func (m *Manager) progressSubscriptions(ctx context.Context) (err error) {
	subscriptions, err := m.dal.GetSubscriptionsNeedingUpdate(ctx)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions needing update: %w", err)
	}
	for _, subscription := range subscriptions {
		if !m.isResponsibleForSubscription(subscription) {
			continue
		}
		logger := log.FromContext(ctx)

		err := m.dal.ProgressSubscription(ctx, subscription)
		if err != nil {
			logger.Errorf(err, "failed to progress subscription")
		}
	}
	return nil
}

func (m *Manager) OnCallCompletion(ctx context.Context, tx *dal.Tx, origin dal.AsyncOriginPubSub, failed bool) error {
	return m.dal.CompleteEventForSubscription(ctx, origin.Subscription.Module, origin.Subscription.Name)
}
