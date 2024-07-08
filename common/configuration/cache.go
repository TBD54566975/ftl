package configuration

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
	"github.com/benbjohnson/clock"
	"github.com/puzpuzpuz/xsync/v3"
)

const (
	syncInitialBackoff  = time.Second * 5
	syncMaxBackoff      = time.Minute * 2
	waitForCacheTimeout = time.Second * 5
)

type updateCacheEvent struct {
	key string
	ref Ref
	// value is nil when the value was deleted
	value optional.Option[[]byte]
}

// cache maintains a cache of providers that can be synced
//
// Loading values always returns the cached value.
// Sync happens periodically.
// Updates do not go through the cache, but the cache is notified after the update occurs.
type cache[R Role] struct {
	providers map[string]*cacheProvider[R]

	topic *pubsub.Topic[updateCacheEvent]
	// used by tests to wait for events to be processed
	topicWaitGroup sync.WaitGroup
}

func newCache[R Role](ctx context.Context, providers []SyncableProvider[R]) *cache[R] {
	cacheProviders := make(map[string]*cacheProvider[R], len(providers))
	for _, provider := range providers {
		cacheProviders[provider.Key()] = &cacheProvider[R]{
			provider: provider,
			values:   xsync.NewMapOf[Ref, SyncedValue](),
			loaded:   make(chan bool),
		}
	}
	cache := &cache[R]{
		providers: cacheProviders,
		topic:     pubsub.New[updateCacheEvent](),
	}
	go cache.sync(ctx, clock.New())

	return cache
}

func (c *cache[R]) load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	providerKey := ProviderKeyForAccessor(key)
	provider, ok := c.providers[key.Scheme]
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
func (c *cache[R]) updatedValue(ref Ref, value []byte, accessor *url.URL) {
	c.topicWaitGroup.Add(1)
	c.topic.Publish(updateCacheEvent{
		key:   ProviderKeyForAccessor(accessor),
		ref:   ref,
		value: optional.Some(value),
	})
}

// deletedValue should be called when a value is deleted in the provider
func (c *cache[R]) deletedValue(ref Ref, pkey string) {
	c.topicWaitGroup.Add(1)
	c.topic.Publish(updateCacheEvent{
		key:   pkey,
		ref:   ref,
		value: optional.None[[]byte](),
	})
}

// sync periodically syncs all syncable providers.
//
// Blocks until the context is cancelled.
// Errors returned by a provider cause retries with exponential backoff.
//
// Events are processed when all providers are not being synced
func (c *cache[R]) sync(ctx context.Context, clock clock.Clock) {
	events := make(chan updateCacheEvent, 64)
	c.topic.Subscribe(events)
	defer c.topic.Unsubscribe(events)

	for {
		select {
		case <-ctx.Done():
			return

		case e := <-events:
			c.processEvent(e)

		// Can not calculate next sync date for each provider as sync intervals can change (eg when follower becomes leader)
		case <-clock.After(time.Second):
			wg := sync.WaitGroup{}
			for _, cp := range c.providers {
				if !cp.needsSync(clock) {
					continue
				}
				wg.Add(1)
				go func(cp *cacheProvider[R]) {
					cp.sync(ctx, clock)
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

// cacheProvider wraps a syncable provider and caches its values.
type cacheProvider[R Role] struct {
	provider SyncableProvider[R]
	values   *xsync.MapOf[Ref, SyncedValue]

	// closed when values have been synced for the first time
	loaded          chan bool
	lastSyncAttempt optional.Option[time.Time]
	currentBackoff  optional.Option[time.Duration]
}

// waitForInitialSync waits until the initial sync has completed.
//
// If values have not yet been synced, this will wait until they are, returning an error if it takes too long.
func (c *cacheProvider[R]) waitForInitialSync() error {
	select {
	case <-time.After(waitForCacheTimeout):
		return fmt.Errorf("%s has not completed sync yet", c.provider.Key())
	case <-c.loaded:
		return nil
	}
}

func (c *cacheProvider[R]) needsSync(clock clock.Clock) bool {
	lastSyncAttempt, ok := c.lastSyncAttempt.Get()
	if !ok {
		return true
	}
	if currentBackoff, ok := c.currentBackoff.Get(); ok {
		return clock.Now().After(lastSyncAttempt.Add(currentBackoff))
	}
	return clock.Now().After(lastSyncAttempt.Add(c.provider.SyncInterval()))
}

func (c *cacheProvider[R]) sync(ctx context.Context, clock clock.Clock) {
	logger := log.FromContext(ctx)

	c.lastSyncAttempt = optional.Some(clock.Now())
	err := c.provider.Sync(ctx, c.values)
	if err != nil {
		logger.Errorf(err, "Error syncing %s", c.provider.Key())
		if backoff, ok := c.currentBackoff.Get(); ok {
			c.currentBackoff = optional.Some(min(syncMaxBackoff, backoff*2))
		} else {
			c.currentBackoff = optional.Some(syncInitialBackoff)
		}
		return
	}
	c.currentBackoff = optional.None[time.Duration]()
	select {
	case <-c.loaded:
		break
	default:
		c.loaded <- true
	}
}

// processEvent updates the cache
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
		c.values.Store(e.ref, SyncedValue{
			Value:        data,
			VersionToken: optional.None[VersionToken](),
		})
	} else {
		// removed value
		c.values.Delete(e.ref)
	}
}
