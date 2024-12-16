package manager

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"

	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/log"
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
	provider *cacheProvider[R]

	// list provider is used to determine which providers are expected to have values, and therefore need to be synced
	listProvider listProvider

	topic *pubsub.Topic[updateCacheEvent]
	// used by tests to wait for events to be processed
	topicWaitGroup *sync.WaitGroup
}

func newCache[R configuration.Role](ctx context.Context, provider optional.Option[configuration.AsynchronousProvider[R]], listProvider listProvider) *cache[R] {
	cache := &cache[R]{
		provider: &cacheProvider[R]{
			provider:   provider,
			values:     map[configuration.Ref]configuration.SyncedValue{},
			loaded:     make(chan bool),
			loadedOnce: &sync.Once{},
		},
		listProvider:   listProvider,
		topic:          pubsub.New[updateCacheEvent](),
		topicWaitGroup: &sync.WaitGroup{},
	}
	go cache.sync(ctx)

	return cache
}

// load is called by the manager to get a value from the cache
func (c *cache[R]) load(ref configuration.Ref, key *url.URL) ([]byte, error) {
	if err := c.provider.waitForInitialSync(); err != nil {
		return nil, err
	}
	value, ok := c.provider.load(ref)
	if !ok {
		return nil, fmt.Errorf("secret not found: %s", ref)
	}
	return value.Value, nil
}

// updatedValue should be called when a value is updated in the provider
func (c *cache[R]) updatedValue(ref configuration.Ref, value []byte, accessor *url.URL) {
	pkey := ProviderKeyForAccessor(accessor)
	provider, ok := c.provider.provider.Get()
	if !ok || provider.Key() != pkey {
		// not syncing this provider
		return
	}
	c.topicWaitGroup.Add(1)
	c.topic.Publish(updateCacheEvent{
		key:   pkey,
		ref:   ref,
		value: optional.Some(value),
	})
}

// deletedValue should be called when a value is deleted in the provider
func (c *cache[R]) deletedValue(ref configuration.Ref, pkey configuration.ProviderKey) {
	provider, ok := c.provider.provider.Get()
	if !ok || provider.Key() != pkey {
		return // provider not set
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
			if !c.provider.needsSync() {
				continue
			}
			c.provider.sync(ctx)
		}
	}
}

func (c *cache[R]) processEvent(e updateCacheEvent) {
	c.provider.processEvent(e)
	// waitGroup updated so testing can wait for events to be processed
	c.topicWaitGroup.Done()
}

// cacheProvider wraps an asynchronous provider and caches its values.
type cacheProvider[R configuration.Role] struct {
	provider   optional.Option[configuration.AsynchronousProvider[R]]
	valuesLock sync.RWMutex
	values     map[configuration.Ref]configuration.SyncedValue

	loaded     chan bool  // closed when values have been synced for the first time
	loadedOnce *sync.Once // ensures we close the loaded channel only once

	lastSyncAttempt optional.Option[time.Time] // updated each time we attempt to sync, regardless of success/failure
	currentBackoff  optional.Option[time.Duration]
}

// waitForInitialSync waits until the initial sync has completed.
//
// If values have not yet been synced, this will wait until they are, returning an error if it takes too long.
func (c *cacheProvider[R]) waitForInitialSync() error {
	provider, ok := c.provider.Get()
	if !ok {
		return nil
	}
	select {
	case <-c.loaded:
		return nil
	case <-time.After(waitForCacheTimeout):
		return fmt.Errorf("%s has not completed sync yet", provider.Key())
	}
}

// needsSync returns true if the provider needs to be synced.
func (c *cacheProvider[R]) needsSync() bool {
	provider, ok := c.provider.Get()
	if !ok {
		return false
	}
	lastSyncAttempt, ok := c.lastSyncAttempt.Get()
	if !ok {
		return true
	}
	if currentBackoff, ok := c.currentBackoff.Get(); ok {
		return time.Now().After(lastSyncAttempt.Add(currentBackoff))
	}
	return time.Now().After(lastSyncAttempt.Add(provider.SyncInterval()))
}

// sync executes sync on the provider and updates the cacheProvider sync state
func (c *cacheProvider[R]) sync(ctx context.Context) {
	provider, ok := c.provider.Get()
	if !ok {
		return
	}
	logger := log.FromContext(ctx)

	c.lastSyncAttempt = optional.Some(time.Now())
	values, err := provider.Sync(ctx)
	if err != nil {
		logger.Errorf(err, "Error syncing %s", provider.Key())
		if backoff, ok := c.currentBackoff.Get(); ok {
			c.currentBackoff = optional.Some(min(syncMaxBackoff, backoff*2))
		} else {
			c.currentBackoff = optional.Some(syncInitialBackoff)
		}
		return
	}
	c.valuesLock.Lock()
	defer c.valuesLock.Unlock()
	c.values = values
	logger.Tracef("Synced provider cache for %s with %d values\n", provider.Key(), len(c.values))
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
	c.valuesLock.Lock()
	defer c.valuesLock.Unlock()
	if data, ok := e.value.Get(); ok {
		// updated value
		c.values[e.ref] = configuration.SyncedValue{
			Value:        data,
			VersionToken: optional.None[configuration.VersionToken](),
		}
	} else {
		// removed value
		delete(c.values, e.ref)
	}
}

func (c *cacheProvider[R]) load(ref configuration.Ref) (configuration.SyncedValue, bool) {
	c.valuesLock.RLock()
	defer c.valuesLock.RUnlock()
	value, ok := c.values[ref]
	return value, ok
}
