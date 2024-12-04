package providers

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/smithy-go"

	"github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/slices"
)

const asmLeaderSyncInterval = time.Minute * 5
const asmTagKey = "ftl"

type asmManager struct {
	client *secretsmanager.Client
}

var _ asmClient = &asmManager{}

func newAsmManager(client *secretsmanager.Client) *asmManager {
	l := &asmManager{
		client: client,
	}
	return l
}

func (l *asmManager) name() string {
	return "asm/leader"
}

// sync retrieves all secrets from ASM and updates the cache
func (l *asmManager) sync(ctx context.Context) (map[configuration.Ref]configuration.SyncedValue, error) {
	// get list of secrets
	refsToLoad := map[configuration.Ref]time.Time{}
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
			return nil, fmt.Errorf("unable to get list of secrets from ASM: %w", err)
		}
		activeSecrets := slices.Filter(out.SecretList, func(s types.SecretListEntry) bool {
			return s.DeletedDate == nil
		})

		for _, s := range activeSecrets {
			ref, err := configuration.ParseRef(*s.Name)
			if err != nil {
				return nil, fmt.Errorf("unable to parse ref from ASM secret: %w", err)
			}
			refsToLoad[ref] = *s.LastChangedDate
		}

		nextToken = optional.Ptr[string](out.NextToken)
		if !nextToken.Ok() {
			break
		}
	}

	values := map[configuration.Ref]configuration.SyncedValue{}
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
			return nil, fmt.Errorf("unable to get batch of secret values from ASM: %w", err)
		}
		for _, s := range out.SecretValues {
			ref, err := configuration.ParseRef(*s.Name)
			if err != nil {
				return nil, fmt.Errorf("unable to parse ref: %w", err)
			}
			// Expect secrets to be strings, not binary
			if s.SecretBinary != nil {
				return nil, fmt.Errorf("secret for %s in ASM is not a string", ref)
			}
			data := unwrapComments([]byte(*s.SecretString))
			values[ref] = configuration.SyncedValue{
				Value:        data,
				VersionToken: optional.Some[configuration.VersionToken](refsToLoad[ref]),
			}
			delete(refsToLoad, ref)
		}
	}
	return nil, nil
}

// store and if the secret already exists, update it.
func (l *asmManager) store(ctx context.Context, ref configuration.Ref, value []byte) (*url.URL, error) {
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

func (l *asmManager) delete(ctx context.Context, ref configuration.Ref) error {
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
