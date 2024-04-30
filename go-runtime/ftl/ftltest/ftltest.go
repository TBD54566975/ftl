// Package ftltest contains test utilities for the ftl package.
package ftltest

import (
	"context"
	"fmt"
	"reflect"

	"github.com/alecthomas/types/optional"

	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/modulecontext"
	"github.com/TBD54566975/ftl/internal/log"
)

// Context suitable for use in testing FTL verbs with provided options
func Context(options ...func(context.Context) error) context.Context {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	context, err := modulecontext.FromEnvironment(ctx, ftl.Module())
	if err != nil {
		panic(err)
	}
	ctx = context.ApplyToContext(ctx)

	mockProvider := newMockVerbProvider()
	ctx = ftl.ApplyCallOverriderToContext(ctx, mockProvider)

	for _, option := range options {
		err = option(ctx)
		if err != nil {
			panic(fmt.Sprintf("error applying option: %v", err))
		}
	}
	return ctx
}

// WithConfig sets a configuration for the current module
//
// To be used when setting up a context for a test:
// ctx := ftltest.Context(
//
//	ftltest.WithConfig(exampleEndpoint, "https://example.com"),
//	... other options
//
// )
func WithConfig[T ftl.ConfigType](config ftl.ConfigValue[T], value T) func(context.Context) error {
	return func(ctx context.Context) error {
		cm := cf.ConfigFromContext(ctx)
		return cm.Set(ctx, cf.Ref{Module: optional.Some(config.Module), Name: config.Name}, value)
	}
}

// WithSecret sets a secret for the current module
//
// To be used when setting up a context for a test:
// ctx := ftltest.Context(
//
//	ftltest.WithSecret(privateKey, "abc123"),
//	... other options
//
// )
func WithSecret[T ftl.SecretType](secret ftl.SecretValue[T], value T) func(context.Context) error {
	return func(ctx context.Context) error {
		sm := cf.SecretsFromContext(ctx)
		return sm.Set(ctx, cf.Ref{Module: optional.Some(secret.Module), Name: secret.Name}, value)
	}
}

// WithDSN sets a DSN for the current module
//
// To be used when setting up a context for a test:
// ctx := ftltest.Context(
//
//	ftltest.WithDSN(dbHandle, "..."),
//	... other options
//
// )
func WithDSN(database ftl.Database, dsn string) func(context.Context) error {
	return func(ctx context.Context) error {
		dbProvider := modulecontext.DBProviderFromContext(ctx)
		return dbProvider.Add(database.Name, database.DBType, dsn)
	}
}

// WhenVerb replaces an implementation for a verb
//
// To be used when setting up a context for a test:
// ctx := ftltest.Context(
//
//	ftltest.WhenVerb(Example.Verb, func(ctx context.Context, req Example.Req) (Example.Resp, error) {
//	    ...
//	}),
//	... other options
//
// )
func WhenVerb[Req any, Resp any](verb ftl.Verb[Req, Resp], fake func(ctx context.Context, req Req) (resp Resp, err error)) func(context.Context) error {
	return func(ctx context.Context) error {
		ref := ftl.CallToRef(verb)
		overrider, ok := ftl.CallOverriderFromContext(ctx)
		if !ok {
			return fmt.Errorf("could not override %v with a fake, context not set up with call overrider", ref)
		}
		mockProvider, ok := overrider.(*mockVerbProvider)
		if !ok {
			return fmt.Errorf("could not override %v with a fake, call overrider is not a MockProvider", ref)
		}
		mockProvider.mocks[ref] = func(ctx context.Context, req any) (resp any, err error) {
			request, ok := req.(Req)
			if !ok {
				return nil, fmt.Errorf("invalid request type %T for %v, expected %v", req, ref, reflect.TypeFor[Req]())
			}
			return fake(ctx, request)
		}
		return nil
	}
}
