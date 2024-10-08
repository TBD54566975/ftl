package manager

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
	"github.com/puzpuzpuz/xsync/v3"

	"github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/slices"
)

const (
	syncInitialBackoff  = time.Second * 1
	syncMaxBackoff      = time.Minute * 2
	waitForCacheTimeout = time.Second * 5
)

type listProvider interface {
	List(ctx context.Context) ([]configuration.Entry, error)
}

type updateCacheEvent struct {
	key configuration.ProviderKey
	ref configuration.Ref
	// value is nil when the value was deleted
	value optional.Option[[]byte]
}

// cache maintains a cache for providers that can be synced
//
// Loading values always returns the cached value.
// Sync happens periodically.
// Updates do not go through the cache, but the cache is notified after the update occurs.
type cache[R configuration.Role] struct {
	providers map[configuration.ProviderKey]*cacheProvider[R]

	// list provider is used to determine which providers are expected to have values, and therefore need to be synced
	listProvider listProvider

	topic *pubsub.Topic[updateCacheEvent]
	// used by tests to wait for events to be processed
	topicWaitGroup *sync.WaitGroup
}

func newCache[R configuration.Role](ctx context.Context, providers []configuration.AsynchronousProvider[R], listProvider listProvider) *cache[R] {
	cacheProviders := make(map[configuration.ProviderKey]*cacheProvider[R], len(providers))
	for _, provider := range providers {
		cacheProviders[provider.Key()] = &cacheProvider[R]{
			provider:   provider,
			values:     xsync.NewMapOf[configuration.Ref, configuration.SyncedValue](),
			loaded:     make(chan bool),
			loadedOnce: &sync.Once{},
		}
	}
	cache := &cache[R]{
		providers:      cacheProviders,
		listProvider:   listProvider,
		topic:          pubsub.New[updateCacheEvent](),
		topicWaitGroup: &sync.WaitGroup{},
	}
	go cache.sync(ctx)

	return cache
}

// load is called by the manager to get a value from the cache
func (c *cache[R]) load(ref configuration.Ref, key *url.URL) ([]byte, error) {
	providerKey := ProviderKeyForAccessor(key)
	provider, ok := c.providers[configuration.ProviderKey(key.Scheme)]
	if !ok {
		return nil, fmt.Errorf("no cache provider for key %q", providerKey)
	}
	if err := provider.waitForInitialSync(); err != nil {
		return nil, err
	}
	value, ok := provider.values.Load(ref)
	if !ok {
		return nil, fmt.Errorf("secret not found: %s", ref)
	}
	return value.Value, nil
}

// updatedValue should be called when a value is updated in the provider
func (c *cache[R]) updatedValue(ref configuration.Ref, value []byte, accessor *url.URL) {
	key := ProviderKeyForAccessor(accessor)
	if _, ok := c.providers[key]; !ok {
		// not syncing this provider
		return
	}
	c.topicWaitGroup.Add(1)
	c.topic.Publish(updateCacheEvent{
		key:   key,
		ref:   ref,
		value: optional.Some(value),
	})
}

// deletedValue should be called when a value is deleted in the provider
func (c *cache[R]) deletedValue(ref configuration.Ref, pkey configuration.ProviderKey) {
	if _, ok := c.providers[pkey]; !ok {
		// not syncing this provider
		return
	}
	c.topicWaitGroup.Add(1)
	c.topic.Publish(updateCacheEvent{
		key:   pkey,
		ref:   ref,
		value: optional.None[[]byte](),
	})
}

// sync periodically syncs all asynchronous providers.
//
// Blocks until the context is cancelled.
// Errors returned by a provider cause retries with exponential backoff.
//
// Events are processed when all providers are not being synced
func (c *cache[R]) sync(ctx context.Context) {
	if len(c.providers) == 0 {
		// nothing to sync
		return
	}

	logger := log.FromContext(ctx)

	events := make(chan updateCacheEvent, 64)
	c.topic.Subscribe(events)
	defer c.topic.Unsubscribe(events)

	// start syncing immediately
	next := time.Now()

	for {
		select {
		case <-ctx.Done():
			return

		case e := <-events:
			c.processEvent(e)

		// Can not calculate next sync date for each provider as sync intervals can change (eg when follower becomes leader)
		case <-time.After(time.Until(next)):
			next = time.Now().Add(time.Second)
			wg := &sync.WaitGroup{}

			providersToSync := []*cacheProvider[R]{}
			for _, cp := range c.providers {
				if cp.needsSync() {
					providersToSync = append(providersToSync, cp)
				}
			}
			if len(providersToSync) == 0 {
				continue
			}
			entries, err := c.listProvider.List(ctx)
			if err != nil {
				logger.Warnf("could not sync: could not get list: %v", err)
				continue
			}
			for _, cp := range providersToSync {
				entriesForProvider := slices.Filter(entries, func(e configuration.Entry) bool {
					return ProviderKeyForAccessor(e.Accessor) == cp.provider.Key()
				})
				wg.Add(1)
				go func(cp *cacheProvider[R]) {
					cp.sync(ctx, entriesForProvider)
					wg.Done()
				}(cp)
			}
			wg.Wait()
		}
	}
}

func (c *cache[R]) processEvent(e updateCacheEvent) {
	if pv, ok := c.providers[e.key]; ok {
		pv.processEvent(e)
	}
	// waitGroup updated so testing can wait for events to be processed
	c.topicWaitGroup.Done()
}

// cacheProvider wraps an asynchronous provider and caches its values.
type cacheProvider[R configuration.Role] struct {
	provider configuration.AsynchronousProvider[R]
	values   *xsync.MapOf[configuration.Ref, configuration.SyncedValue]

	loaded     chan bool  // closed when values have been synced for the first time
	loadedOnce *sync.Once // ensures we close the loaded channel only once

	lastSyncAttempt optional.Option[time.Time] // updated each time we attempt to sync, regardless of success/failure
	currentBackoff  optional.Option[time.Duration]
}

// waitForInitialSync waits until the initial sync has completed.
//
// If values have not yet been synced, this will wait until they are, returning an error if it takes too long.
func (c *cacheProvider[R]) waitForInitialSync() error {
	select {
	case <-c.loaded:
		return nil
	case <-time.After(waitForCacheTimeout):
		return fmt.Errorf("%s has not completed sync yet", c.provider.Key())
	}
}

// needsSync returns true if the provider needs to be synced.
func (c *cacheProvider[R]) needsSync() bool {
	lastSyncAttempt, ok := c.lastSyncAttempt.Get()
	if !ok {
		return true
	}
	if currentBackoff, ok := c.currentBackoff.Get(); ok {
		return time.Now().After(lastSyncAttempt.Add(currentBackoff))
	}
	return time.Now().After(lastSyncAttempt.Add(c.provider.SyncInterval()))
}

// sync executes sync on the provider and updates the cacheProvider sync state
func (c *cacheProvider[R]) sync(ctx context.Context, entries []configuration.Entry) {
	logger := log.FromContext(ctx)

	c.lastSyncAttempt = optional.Some(time.Now())
	err := c.provider.Sync(ctx, entries, c.values)
	if err != nil {
		logger.Errorf(err, "Error syncing %s", c.provider.Key())
		if backoff, ok := c.currentBackoff.Get(); ok {
			c.currentBackoff = optional.Some(min(syncMaxBackoff, backoff*2))
		} else {
			c.currentBackoff = optional.Some(syncInitialBackoff)
		}
		return
	}
	logger.Tracef("Synced provider cache for %s with %d values\n", c.provider.Key(), c.values.Size())
	c.currentBackoff = optional.None[time.Duration]()
	c.loadedOnce.Do(func() {
		close(c.loaded)
	})
}

// processEvent updates the cache after a value was set or deleted
func (c *cacheProvider[R]) processEvent(e updateCacheEvent) {
	select {
	case <-c.loaded:
		break
	default:
		// skip event if initial sync has not successfully completed
		return
	}
	if data, ok := e.value.Get(); ok {
		// updated value
		c.values.Store(e.ref, configuration.SyncedValue{
			Value:        data,
			VersionToken: optional.None[configuration.VersionToken](),
		})
	} else {
		// removed value
		c.values.Delete(e.ref)
	}
}
