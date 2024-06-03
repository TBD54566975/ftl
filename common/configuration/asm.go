package configuration

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/TBD54566975/ftl/internal/slices"

	. "github.com/alecthomas/types/optional" //nolint:stylecheck
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/aws/smithy-go"
)

// ASM implements Resolver and Provider for AWS Secrets Manager (ASM).
//
// The resolver does a direct/proxy map from a Ref to a URL, module.name <-> asm://module.name and does not access ASM at all.
type ASM struct {
	Client secretsmanager.Client
}

var _ Resolver[Secrets] = &ASM{}
var _ Provider[Secrets] = &ASM{}

func asmURLForRef(ref Ref) *url.URL {
	return &url.URL{
		Scheme: "asm",
		Host:   ref.String(),
	}
}

func (ASM) Role() Secrets {
	return Secrets{}
}

func (ASM) Key() string {
	return "asm"
}

func (ASM) Get(ctx context.Context, ref Ref) (*url.URL, error) {
	return asmURLForRef(ref), nil
}

func (ASM) Set(ctx context.Context, ref Ref, key *url.URL) error {
	expectedKey := asmURLForRef(ref)
	if key.String() != expectedKey.String() {
		return fmt.Errorf("key does not match expected key for ref: %s", expectedKey)
	}

	return nil
}

// Unset does nothing because this resolver does not record any state.
func (ASM) Unset(ctx context.Context, ref Ref) error {
	return nil
}

// List all secrets in the account. This might require multiple calls to the AWS API if there are more than 100 secrets.
func (a ASM) List(ctx context.Context) ([]Entry, error) {
	nextToken := None[string]()
	entries := []Entry{}
	for {
		out, err := a.Client.ListSecrets(ctx, &secretsmanager.ListSecretsInput{
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
func (a ASM) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	expectedKey := asmURLForRef(ref)
	if key.String() != expectedKey.String() {
		return nil, fmt.Errorf("key does not match expected key for ref: %s", expectedKey)
	}

	out, err := a.Client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
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

// Store and if the secret already exists, update it.
func (a ASM) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	_, err := a.Client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(ref.String()),
		SecretString: aws.String(string(value)),
	})

	// https://github.com/aws/aws-sdk-go-v2/issues/1110#issuecomment-1054643716
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "ResourceExistsException" {
		_, err = a.Client.UpdateSecret(ctx, &secretsmanager.UpdateSecretInput{
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

func (a ASM) Delete(ctx context.Context, ref Ref) error {
	var t = true
	_, err := a.Client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(ref.String()),
		ForceDeleteWithoutRecovery: &t,
	})
	if err != nil {
		return fmt.Errorf("unable to delete secret: %w", err)
	}

	return nil
}
