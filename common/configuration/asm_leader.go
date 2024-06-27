package configuration

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/puzpuzpuz/xsync/v3"

	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/alecthomas/types/optional"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/smithy-go"
)

const asmLeaderSyncInterval = time.Minute * 5

type asmLeader struct {
	client *secretsmanager.Client
	cache  *secretsCache
}

var _ asmClient = &asmLeader{}

func newASMLeader(ctx context.Context, client *secretsmanager.Client, clock clock.Clock) *asmLeader {
	l := &asmLeader{
		client: client,
		cache:  newSecretsCache("asm-leader"),
	}
	go l.cache.sync(ctx, asmLeaderSyncInterval, func(ctx context.Context, secrets *xsync.MapOf[Ref, cachedSecret]) error {
		return l.sync(ctx, secrets)
	}, clock)
	return l
}

// sync retrieves all secrets from ASM and updates the cache
func (l *asmLeader) sync(ctx context.Context, secrets *xsync.MapOf[Ref, cachedSecret]) error {
	previous := map[Ref]cachedSecret{}
	secrets.Range(func(ref Ref, value cachedSecret) bool {
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
			Filters: []types.Filter{
				{Key: types.FilterNameStringTypeTagKey, Values: []string{"_ftl"}},
			},
		})
		if err != nil {
			return fmt.Errorf("unable to get list of secrets from ASM: %w", err)
		}

		var activeSecrets = slices.Filter(out.SecretList, func(s types.SecretListEntry) bool {
			return s.DeletedDate == nil
		})
		for _, s := range activeSecrets {
			ref, err := ParseRef(*s.Name)
			if err != nil {
				return fmt.Errorf("unable to parse ref from ASM secret: %w", err)
			}
			seen[ref] = true

			// check if we already have the value from previous sync
			if pValue, ok := previous[ref]; ok && pValue.versionToken == optional.Some[any](*s.LastChangedDate) {
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
			secrets.Delete(ref)
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
			Filters: []types.Filter{
				{Key: types.FilterNameStringTypeName, Values: secretIDs},
				{Key: types.FilterNameStringTypeTagKey, Values: []string{"_ftl"}},
			},
		})
		if err != nil {
			return fmt.Errorf("unable to get batch of secret values from ASM: %w", err)
		}
		for _, s := range out.SecretValues {
			ref, err := ParseRef(*s.Name)
			if err != nil {
				return fmt.Errorf("unable to parse ref: %w", err)
			}
			// Expect secrets to be strings, not binary
			if s.SecretBinary != nil {
				return fmt.Errorf("secret for %s in ASM is not a string", ref)
			}
			data := []byte(*s.SecretString)
			secrets.Store(ref, cachedSecret{
				value:        data,
				versionToken: optional.Some[any](refsToLoad[ref]),
			})
			delete(refsToLoad, ref)
		}
	}
	return nil
}

// list all secrets in the account.
func (l *asmLeader) list(ctx context.Context) ([]Entry, error) {
	entries := []Entry{}
	err := l.cache.iterate(func(ref Ref, _ []byte) {
		entries = append(entries, Entry{
			Ref:      ref,
			Accessor: asmURLForRef(ref),
		})
	})
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func (l *asmLeader) load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	return l.cache.getSecret(ref)
}

// store and if the secret already exists, update it.
func (l *asmLeader) store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	_, err := l.client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(ref.String()),
		SecretString: aws.String(string(value)),
		Tags: []types.Tag{
			{Key: aws.String("_ftl"), Value: aws.String(ref.Module.Default("_"))},
		},
	})

	// https://github.com/aws/aws-sdk-go-v2/issues/1110#issuecomment-1054643716
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "ResourceExistsException" {
		_, err = l.client.UpdateSecret(ctx, &secretsmanager.UpdateSecretInput{
			SecretId:     aws.String(ref.String()),
			SecretString: aws.String(string(value)),
		})
		if err != nil {
			return nil, fmt.Errorf("unable to update secret in ASM: %w", err)
		}

	} else if err != nil {
		return nil, fmt.Errorf("unable to store secret in ASM: %w", err)
	}
	l.cache.updatedSecret(ref, value)
	return asmURLForRef(ref), nil
}

func (l *asmLeader) delete(ctx context.Context, ref Ref) error {
	var t = true
	_, err := l.client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(ref.String()),
		ForceDeleteWithoutRecovery: &t,
	})
	if err != nil {
		return fmt.Errorf("unable to delete secret from ASM: %w", err)
	}
	l.cache.deletedSecret(ref)
	return nil
}
