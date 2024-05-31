package configuration

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"github.com/TBD54566975/ftl/internal/slices"

	"github.com/alecthomas/types/optional"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/smithy-go"
)

// ASM implements Resolver and Provider for AWS Secrets Manager (ASM).
//
// The resolver does a direct/proxy map from a Ref to a URL, module.name <-> asm://module.name and does not access ASM at all.
type ASM[R Secrets] struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Endpoint        optional.Option[string]
}

var _ Resolver[Secrets] = ASM[Secrets]{}
var _ Provider[Secrets] = ASM[Secrets]{}
var _ MutableProvider[Secrets] = ASM[Secrets]{}

var (
	asmOnce   sync.Once
	asmClient *secretsmanager.Client
	errClient error
)

func asmURLForRef(ref Ref) *url.URL {
	return &url.URL{
		Scheme: "asm",
		Host:   ref.String(),
	}
}

func (a ASM[R]) client(ctx context.Context) (*secretsmanager.Client, error) {
	asmOnce.Do(func() {
		var optFns []func(*config.LoadOptions) error

		// Use a static credentials provider if access key and secret are provided.
		// Otherwise, the SDK will use the default credential chain (env vars, iam, etc).
		if a.AccessKeyID != "" {
			credentials := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(a.AccessKeyID, a.SecretAccessKey, ""))
			optFns = append(optFns, config.WithCredentialsProvider(credentials))
		}

		if a.Region != "" {
			optFns = append(optFns, config.WithRegion(a.Region))
		}

		cfg, err := config.LoadDefaultConfig(ctx, optFns...)
		if err != nil {
			errClient = fmt.Errorf("unable to load aws config: %w", err)
			return
		}

		asmClient = secretsmanager.NewFromConfig(cfg, func(o *secretsmanager.Options) {
			e, ok := a.Endpoint.Get()
			if ok {
				o.BaseEndpoint = aws.String(e)
			}
		})

	})

	return asmClient, errClient
}

func (a ASM[R]) Role() R {
	var r R
	return r
}

func (a ASM[R]) Key() string {
	return "asm"
}

func (a ASM[R]) Get(ctx context.Context, ref Ref) (*url.URL, error) {
	return asmURLForRef(ref), nil
}

func (a ASM[R]) Set(ctx context.Context, ref Ref, key *url.URL) error {
	expectedKey := asmURLForRef(ref)
	if key.String() != expectedKey.String() {
		return fmt.Errorf("key does not match expected key for ref: %s", expectedKey)
	}

	return nil
}

// Unset does nothing because this resolver does not record any state.
func (a ASM[R]) Unset(ctx context.Context, ref Ref) error {
	return nil
}

// List all secrets in the account. This might require multiple calls to the AWS API if there are more than 100 secrets.
func (a ASM[R]) List(ctx context.Context) ([]Entry, error) {
	c, err := a.client(ctx)
	if err != nil {
		return nil, err
	}

	nextToken := optional.None[string]()
	entries := []Entry{}
	for {
		out, err := c.ListSecrets(ctx, &secretsmanager.ListSecretsInput{
			MaxResults: aws.Int32(100),
			NextToken:  nextToken.Ptr(),
		})
		if err != nil {
			return nil, fmt.Errorf("unable to list secrets: %w", err)
		}

		var activeSecrets = slices.Filter(out.SecretList, func(s types.SecretListEntry) bool {
			return s.DeletedDate == nil
		})
		page, err := slices.MapErr(activeSecrets, func(s types.SecretListEntry) (Entry, error) {
			var ref Ref
			ref, err = ParseRef(*s.Name)
			if err != nil {
				return Entry{}, fmt.Errorf("unable to parse ref: %w", err)
			}

			return Entry{
				Ref:      ref,
				Accessor: asmURLForRef(ref),
			}, nil
		})
		if err != nil {
			return nil, err
		}

		entries = append(entries, page...)

		nextToken = optional.Ptr[string](out.NextToken)
		if !nextToken.Ok() {
			break
		}
	}

	return entries, nil
}

// Load only supports loading "string" secrets, not binary secrets.
func (a ASM[R]) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	expectedKey := asmURLForRef(ref)
	if key.String() != expectedKey.String() {
		return nil, fmt.Errorf("key does not match expected key for ref: %s", expectedKey)
	}

	c, err := a.client(ctx)
	if err != nil {
		return nil, err
	}

	out, err := c.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(ref.String()),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve secret: %w", err)
	}

	// Secret is a string
	if out.SecretBinary != nil {
		return nil, fmt.Errorf("secret is not a string")
	}

	return []byte(*out.SecretString), nil
}

func (a ASM[R]) Writer() bool {
	return true
}

// Store and if the secret already exists, update it.
func (a ASM[R]) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	c, err := a.client(ctx)
	if err != nil {
		return nil, err
	}

	_, err = c.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(ref.String()),
		SecretString: aws.String(string(value)),
	})

	// https://github.com/aws/aws-sdk-go-v2/issues/1110#issuecomment-1054643716
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "ResourceExistsException" {
		_, err = c.UpdateSecret(ctx, &secretsmanager.UpdateSecretInput{
			SecretId:     aws.String(ref.String()),
			SecretString: aws.String(string(value)),
		})
		if err != nil {
			return nil, fmt.Errorf("unable to update secret: %w", err)
		}

	} else if err != nil {
		return nil, fmt.Errorf("unable to store secret: %w", err)
	}

	return asmURLForRef(ref), nil
}

func (a ASM[R]) Delete(ctx context.Context, ref Ref) error {
	c, err := a.client(ctx)
	if err != nil {
		return err
	}

	var t = true
	_, err = c.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(ref.String()),
		ForceDeleteWithoutRecovery: &t,
	})
	if err != nil {
		return fmt.Errorf("unable to delete secret: %w", err)
	}

	return nil
}
