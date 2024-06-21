package configuration

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/puzpuzpuz/xsync/v3"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/alecthomas/types/optional"
	"github.com/alecthomas/types/pubsub"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/smithy-go"
)

const (
	syncInterval       = time.Minute * 5
	syncInitialBackoff = time.Second * 5
	loadSecretsTimeout = time.Second * 5
)

type asmSecretValue struct {
	value []byte

	// lastModified is retrieved from ASM when syncing.
	// it is nil when our cache is updated after writing to ASM (lastModified is not returned when writing),
	// until the next sync refetches it from ASM
	lastModified optional.Option[time.Time]
}

type updateASMSecretEvent struct {
	ref Ref
	// value is nil when the secret was deleted
	value optional.Option[[]byte]
}

type asmLeader struct {
	client *secretsmanager.Client

	// indicates if the initial sync with ASM has finished
	loaded chan bool

	// secrets is a map of secrets that have been loaded from ASM.
	// optional is nil when not loaded yet
	// it is updated via:
	// - polling ASM
	// - when we write to ASM
	secrets *xsync.MapOf[Ref, asmSecretValue]

	topic *pubsub.Topic[updateASMSecretEvent]
	// used by tests to wait for events to be processed
	topicWaitGroup sync.WaitGroup
}

var _ asmClient = &asmLeader{}

func newASMLeader(ctx context.Context, client *secretsmanager.Client, clock clock.Clock) *asmLeader {
	l := &asmLeader{
		client:  client,
		secrets: xsync.NewMapOf[Ref, asmSecretValue](),
		topic:   pubsub.New[updateASMSecretEvent](),
		loaded:  make(chan bool),
	}
	go func() {
		l.watchForUpdates(ctx, clock)
	}()
	return l
}

func (l *asmLeader) watchForUpdates(ctx context.Context, clock clock.Clock) {
	logger := log.FromContext(ctx)
	events := make(chan updateASMSecretEvent, 64)
	l.topic.Subscribe(events)
	defer l.topic.Unsubscribe(events)

	nextSync := clock.Now()
	backOff := syncInitialBackoff
	loaded := false
	for {
		select {
		case <-ctx.Done():
			return

		case e := <-events:
			if loaded {
				l.processEvent(e)
			}

		case <-clock.After(clock.Until(nextSync)):
			nextSync = clock.Now().Add(syncInterval)

			err := l.sync(ctx)
			if err == nil {
				backOff = syncInitialBackoff
				if !loaded {
					loaded = true
					close(l.loaded)
				}
				continue
			}
			// back off if we fail to sync
			logger.Warnf("unable to sync ASM secrets: %v", err)
			nextSync = clock.Now().Add(backOff)
			if nextSync.After(clock.Now().Add(syncInterval)) {
				nextSync = clock.Now().Add(syncInterval)
			} else {
				backOff *= 2
			}
		}
	}
}

// sync retrieves all secrets from ASM and updates the cache
func (l *asmLeader) sync(ctx context.Context) error {
	previous := map[Ref]asmSecretValue{}
	l.secrets.Range(func(ref Ref, value asmSecretValue) bool {
		previous[ref] = value
		return true
	})
	seen := map[Ref]bool{}

	// get list of secrets
	refsToLoad := map[Ref]time.Time{}
	nextToken := optional.None[string]()
	for {
		out, err := l.client.ListSecrets(ctx, &secretsmanager.ListSecretsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nextToken.Ptr(),
		})
		if err != nil {
			return fmt.Errorf("unable to list secrets: %w", err)
		}

		var activeSecrets = slices.Filter(out.SecretList, func(s types.SecretListEntry) bool {
			return s.DeletedDate == nil
		})
		for _, s := range activeSecrets {
			ref, err := ParseRef(*s.Name)
			if err != nil {
				return fmt.Errorf("unable to parse ref: %w", err)
			}
			seen[ref] = true

			// check if we already have the value from previous sync
			if pValue, ok := previous[ref]; ok && pValue.lastModified == optional.Some(*s.LastChangedDate) {
				continue
			}
			refsToLoad[ref] = *s.LastChangedDate
		}

		nextToken = optional.Ptr[string](out.NextToken)
		if !nextToken.Ok() {
			break
		}
	}

	// remove secrets not found in ASM
	for ref := range previous {
		if _, ok := seen[ref]; !ok {
			l.secrets.Delete(ref)
		}
	}

	// get values for new and updated secrets
	for len(refsToLoad) > 0 {
		batchSize := 20
		secretIDs := []string{}
		for ref := range refsToLoad {
			secretIDs = append(secretIDs, ref.String())
			if len(secretIDs) == batchSize {
				break
			}
		}
		out, err := l.client.BatchGetSecretValue(ctx, &secretsmanager.BatchGetSecretValueInput{
			SecretIdList: secretIDs,
		})
		if err != nil {
			return fmt.Errorf("unable to batch get secret values: %w", err)
		}
		for _, s := range out.SecretValues {
			ref, err := ParseRef(*s.Name)
			if err != nil {
				return fmt.Errorf("unable to parse ref: %w", err)
			}
			// Expect secrets to be strings, not binary
			if s.SecretBinary != nil {
				return fmt.Errorf("secret for %s is not a string", ref)
			}
			data := []byte(*s.SecretString)
			l.secrets.Store(ref, asmSecretValue{
				value:        data,
				lastModified: optional.Some(refsToLoad[ref]),
			})
			delete(refsToLoad, ref)
		}
	}
	return nil
}

// processEvent updates the cache after writes to ASM
func (l *asmLeader) processEvent(e updateASMSecretEvent) {
	// waitGroup updated so testing can wait for events to be processed
	defer l.topicWaitGroup.Done()

	if data, ok := e.value.Get(); ok {
		// updated value
		l.secrets.Store(e.ref, asmSecretValue{
			value:        data,
			lastModified: optional.None[time.Time](),
		})
	} else {
		// removed value
		l.secrets.Delete(e.ref)
	}
}

// publishes an event to update the cache
//
// Called after writing to ASM.
// Wrapped in a method to ensure we always update topicWaitGroup
func (l *asmLeader) publish(event updateASMSecretEvent) {
	l.topicWaitGroup.Add(1)
	l.topic.Publish(event)
}

// waitForSecrets waits until the initial sync of secrets has completed.
//
// If secrets have not yet been synced, this will retry until they are, returning an error if it takes too long.
func (l *asmLeader) waitForSecrets() error {
	select {
	case <-time.After(loadSecretsTimeout):
		return errors.New("secrets not synced from ASM yet")
	case <-l.loaded:
		return nil
	}
}

// list all secrets in the account.
func (l *asmLeader) list(ctx context.Context) ([]Entry, error) {
	if err := l.waitForSecrets(); err != nil {
		return nil, err
	}
	entries := []Entry{}
	l.secrets.Range(func(ref Ref, value asmSecretValue) bool {
		entries = append(entries, Entry{
			Ref:      ref,
			Accessor: asmURLForRef(ref),
		})
		return true
	})
	return entries, nil
}

func (l *asmLeader) load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	if err := l.waitForSecrets(); err != nil {
		return nil, err
	}
	if v, ok := l.secrets.Load(ref); ok {
		return v.value, nil
	}
	return nil, fmt.Errorf("secret not found: %s", ref)
}

// store and if the secret already exists, update it.
func (l *asmLeader) store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	_, err := l.client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(ref.String()),
		SecretString: aws.String(string(value)),
	})

	// https://github.com/aws/aws-sdk-go-v2/issues/1110#issuecomment-1054643716
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "ResourceExistsException" {
		_, err = l.client.UpdateSecret(ctx, &secretsmanager.UpdateSecretInput{
			SecretId:     aws.String(ref.String()),
			SecretString: aws.String(string(value)),
		})
		if err != nil {
			return nil, fmt.Errorf("unable to update secret: %w", err)
		}

	} else if err != nil {
		return nil, fmt.Errorf("unable to store secret: %w", err)
	}
	l.publish(updateASMSecretEvent{
		ref:   ref,
		value: optional.Some(value),
	})
	return asmURLForRef(ref), nil
}

func (l *asmLeader) delete(ctx context.Context, ref Ref) error {
	var t = true
	_, err := l.client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(ref.String()),
		ForceDeleteWithoutRecovery: &t,
	})
	if err != nil {
		return fmt.Errorf("unable to delete secret: %w", err)
	}
	l.publish(updateASMSecretEvent{
		ref:   ref,
		value: optional.None[[]byte](),
	})
	return nil
}
