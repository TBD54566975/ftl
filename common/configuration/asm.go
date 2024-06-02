package configuration

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/TBD54566975/ftl/internal/slices"

	. "github.com/alecthomas/types/optional" //nolint:stylecheck
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
type ASM struct {
	client secretsmanager.Client
}

var _ Resolver[Secrets] = &ASM{}
var _ Provider[Secrets] = &ASM{}
var _ MutableProvider[Secrets] = &ASM{}

func asmURLForRef(ref Ref) *url.URL {
	return &url.URL{
		Scheme: "asm",
		Host:   ref.String(),
	}
}

// NewASMWithDefaultCredentials creates a new ASM resolver/provider that uses the default AWS SDK credentials chain.
func NewASMWithDefaultCredentials(ctx context.Context) (*ASM, error) {
	return NewASM(ctx, None[string](), None[string](), None[string](), None[string]())
}

// NewASM creates a new ASM resolver/provider with optional access key, secret access key, region, and endpoint.
func NewASM(ctx context.Context, accessKeyID, secretAccessKey, region, endpoint Option[string]) (*ASM, error) {
	var optFns []func(*config.LoadOptions) error

	// Use a static credentials provider if access key and secret are provided.
	// Otherwise, the SDK will use the default credential chain (env vars, iam, etc).
	if access, oka := accessKeyID.Get(); oka {
		secret, oks := secretAccessKey.Get()
		if !oks {
			return nil, errors.New("secret access key must be provided if access key ID is provided")
		}
		cc := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(access, secret, ""))
		optFns = append(optFns, config.WithCredentialsProvider(cc))
	}

	if r, ok := region.Get(); ok {
		optFns = append(optFns, config.WithRegion(r))
	}

	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, fmt.Errorf("unable to load aws config: %w", err)
	}

	asmClient := secretsmanager.NewFromConfig(cfg, func(o *secretsmanager.Options) {
		e, ok := endpoint.Get()
		if ok {
			o.BaseEndpoint = aws.String(e)
		}
	})
	asm := ASM{
		client: *asmClient,
	}
	return &asm, nil
}

func (a *ASM) Role() Secrets {
	return Secrets{}
}

func (a *ASM) Key() string {
	return "asm"
}

func (a *ASM) Get(ctx context.Context, ref Ref) (*url.URL, error) {
	return asmURLForRef(ref), nil
}

func (a *ASM) Set(ctx context.Context, ref Ref, key *url.URL) error {
	expectedKey := asmURLForRef(ref)
	if key.String() != expectedKey.String() {
		return fmt.Errorf("key does not match expected key for ref: %s", expectedKey)
	}

	return nil
}

// Unset does nothing because this resolver does not record any state.
func (a *ASM) Unset(ctx context.Context, ref Ref) error {
	return nil
}

// List all secrets in the account. This might require multiple calls to the AWS API if there are more than 100 secrets.
func (a *ASM) List(ctx context.Context) ([]Entry, error) {
	nextToken := None[string]()
	entries := []Entry{}
	for {
		out, err := a.client.ListSecrets(ctx, &secretsmanager.ListSecretsInput{
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

		nextToken = Ptr[string](out.NextToken)
		if !nextToken.Ok() {
			break
		}
	}

	return entries, nil
}

// Load only supports loading "string" secrets, not binary secrets.
func (a *ASM) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	expectedKey := asmURLForRef(ref)
	if key.String() != expectedKey.String() {
		return nil, fmt.Errorf("key does not match expected key for ref: %s", expectedKey)
	}

	out, err := a.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
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

func (a *ASM) Writer() bool {
	return true
}

// Store and if the secret already exists, update it.
func (a *ASM) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	_, err := a.client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(ref.String()),
		SecretString: aws.String(string(value)),
	})

	// https://github.com/aws/aws-sdk-go-v2/issues/1110#issuecomment-1054643716
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "ResourceExistsException" {
		_, err = a.client.UpdateSecret(ctx, &secretsmanager.UpdateSecretInput{
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

func (a *ASM) Delete(ctx context.Context, ref Ref) error {
	var t = true
	_, err := a.client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(ref.String()),
		ForceDeleteWithoutRecovery: &t,
	})
	if err != nil {
		return fmt.Errorf("unable to delete secret: %w", err)
	}

	return nil
}
