package configuration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
	"github.com/benbjohnson/clock"
	"github.com/puzpuzpuz/xsync/v3"
)

const (
	syncInitialBackoff = time.Second * 5
	loadSecretsTimeout = time.Second * 5
)

type cachedSecret struct {
	value []byte

	// versionToken is a way of storing a version provided by the source of truth (eg: lastModified)
	// it is nil when:
	// - the owner of the cache is not using version tokens
	// - the cache is updated after writing
	versionToken optional.Option[any]
}

type updatedSecretEvent struct {
	ref Ref
	// value is nil when the secret was deleted
	value optional.Option[[]byte]
}

type secretsCache struct {
	name string

	// closed when secrets have been synced for the first time
	loaded chan bool

	// secrets is a map of secrets that have been synced.
	// optional is nil when not loaded yet
	// it is updated via:
	// - periodic syncs
	// - when we update or delete a secret
	secrets *xsync.MapOf[Ref, cachedSecret]

	topic *pubsub.Topic[updatedSecretEvent]
	// used by tests to wait for events to be processed
	topicWaitGroup sync.WaitGroup
}

func newSecretsCache(name string) *secretsCache {
	return &secretsCache{
		name:    name,
		loaded:  make(chan bool),
		secrets: xsync.NewMapOf[Ref, cachedSecret](),
		topic:   pubsub.New[updatedSecretEvent](),
	}
}

func (c *secretsCache) getSecret(ref Ref) ([]byte, error) {
	if err := c.waitForSecrets(); err != nil {
		return nil, err
	}
	result, ok := c.secrets.Load(ref)
	if !ok {
		return nil, fmt.Errorf("secret not found: %s", ref)
	}
	return result.value, nil
}

func (c *secretsCache) iterate(f func(ref Ref, value []byte)) error {
	if err := c.waitForSecrets(); err != nil {
		return err
	}
	c.secrets.Range(func(ref Ref, value cachedSecret) bool {
		f(ref, value.value)
		return true
	})
	return nil
}

func (c *secretsCache) updatedSecret(ref Ref, value []byte) {
	c.topicWaitGroup.Add(1)
	c.topic.Publish(updatedSecretEvent{
		ref:   ref,
		value: optional.Some(value),
	})
}

func (c *secretsCache) deletedSecret(ref Ref) {
	c.topicWaitGroup.Add(1)
	c.topic.Publish(updatedSecretEvent{
		ref:   ref,
		value: optional.None[[]byte](),
	})
}

// sync is used to update the secrets cache
//
// Blocks until the context is cancelled.
// Synchronizer is a function that will be called to sync the secrets cache
// Errors returned by synchronizer cause retries with exponential backoff.
func (c *secretsCache) sync(ctx context.Context, frequency time.Duration, synchronizer func(ctx context.Context, secrets *xsync.MapOf[Ref, cachedSecret]) error, clock clock.Clock) {
	logger := log.FromContext(ctx)
	events := make(chan updatedSecretEvent, 64)
	c.topic.Subscribe(events)
	defer c.topic.Unsubscribe(events)

	nextSync := clock.Now()
	backOff := syncInitialBackoff
	loaded := false
	for {
		select {
		case <-ctx.Done():
			return

		case e := <-events:
			if loaded {
				c.processEvent(e)
			}

		case <-clock.After(clock.Until(nextSync)):
			nextSync = clock.Now().Add(frequency)

			err := synchronizer(ctx, c.secrets)
			if err == nil {
				backOff = syncInitialBackoff
				if !loaded {
					loaded = true
					close(c.loaded)
				}
				continue
			}
			// back off if we fail to sync
			logger.Warnf("Unable to sync %s: %v", c.name, err)
			nextSync = clock.Now().Add(backOff)
			if nextSync.After(clock.Now().Add(frequency)) {
				nextSync = clock.Now().Add(frequency)
			} else {
				backOff *= 2
			}
		}
	}
}

// processEvent updates the cache after updating the secret
func (c *secretsCache) processEvent(e updatedSecretEvent) {
	// waitGroup updated so testing can wait for events to be processed
	defer c.topicWaitGroup.Done()

	if data, ok := e.value.Get(); ok {
		// updated value
		c.secrets.Store(e.ref, cachedSecret{
			value:        data,
			versionToken: optional.None[any](),
		})
	} else {
		// removed value
		c.secrets.Delete(e.ref)
	}
}

// waitForSecrets waits until the initial sync of secrets has completed.
//
// If secrets have not yet been synced, this will retry until they are, returning an error if it takes too long.
func (c *secretsCache) waitForSecrets() error {
	select {
	case <-time.After(loadSecretsTimeout):
		return fmt.Errorf("secrets not synced for %s yet", c.name)
	case <-c.loaded:
		return nil
	}
}
