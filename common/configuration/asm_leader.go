package configuration

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/puzpuzpuz/xsync/v3"

	"github.com/TBD54566975/ftl/internal/slices"
	"github.com/alecthomas/types/optional"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/smithy-go"
)

const asmLeaderSyncInterval = time.Minute * 5
const asmTagKey = "ftl"

type asmLeader struct {
	client *secretsmanager.Client
}

var _ asmClient = &asmLeader{}

func newASMLeader(ctx context.Context, client *secretsmanager.Client) *asmLeader {
	l := &asmLeader{
		client: client,
	}
	return l
}

func (l *asmLeader) syncInterval() time.Duration {
	return asmLeaderSyncInterval
}

// sync retrieves all secrets from ASM and updates the cache
func (l *asmLeader) sync(ctx context.Context, values *xsync.MapOf[Ref, SyncedValue]) error {
	previous := map[Ref]SyncedValue{}
	values.Range(func(ref Ref, value SyncedValue) bool {
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
				{Key: types.FilterNameStringTypeTagKey, Values: []string{asmTagKey}},
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
			if pValue, ok := previous[ref]; ok && pValue.VersionToken == optional.Some[VersionToken](*s.LastChangedDate) {
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
			values.Delete(ref)
		}
	}

	// get values for new and updated secrets
	for len(refsToLoad) > 0 {
		// ASM returns an error when there are more than 10 filters
		// A batch size of 9 + 1 tag filter keeps us within this limit
		batchSize := 9
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
				{Key: types.FilterNameStringTypeTagKey, Values: []string{asmTagKey}},
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
			data := unwrapComments([]byte(*s.SecretString))
			values.Store(ref, SyncedValue{
				Value:        data,
				VersionToken: optional.Some[VersionToken](refsToLoad[ref]),
			})
			delete(refsToLoad, ref)
		}
	}
	return nil
}

// store and if the secret already exists, update it.
func (l *asmLeader) store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	valueWithComments := aws.String(string(wrapWithComments(value, defaultSecretModificationWarning)))
	_, err := l.client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(ref.String()),
		SecretString: valueWithComments,
		Tags: []types.Tag{
			{Key: aws.String(asmTagKey), Value: aws.String(ref.Module.Default(""))},
		},
	})

	// https://github.com/aws/aws-sdk-go-v2/issues/1110#issuecomment-1054643716
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "ResourceExistsException" {
		_, err = l.client.UpdateSecret(ctx, &secretsmanager.UpdateSecretInput{
			SecretId:     aws.String(ref.String()),
			SecretString: valueWithComments,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to update secret in ASM: %w", err)
		}

	} else if err != nil {
		return nil, fmt.Errorf("unable to store secret in ASM: %w", err)
	}
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
	return nil
}
