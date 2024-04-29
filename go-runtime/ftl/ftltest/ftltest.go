// Package ftltest contains test utilities for the ftl package.
package ftltest

import (
	"context"
	"fmt"
	"reflect"

	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/modulecontext"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/alecthomas/types/optional"
)

type DBType int32

const (
	DBTypePostgres = DBType(modulecontext.DBTypePostgres)
)

// Context suitable for use in testing FTL verbs with provided options
func Context(options ...func(context.Context) error) context.Context {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	context, err := modulecontext.FromEnvironment(ctx, ftl.Module())
	if err != nil {
		panic(err)
	}
	ctx = context.ApplyToContext(ctx)

	mockProvider := newMockProvider()
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
// To be used with Context(...)
func WithConfig(name string, value any) func(context.Context) error {
	return func(ctx context.Context) error {
		cm := cf.ConfigFromContext(ctx)
		return cm.Set(ctx, cf.Ref{Module: optional.Some(ftl.Module()), Name: name}, value)
	}
}

// WithSecret sets a secret for the current module
//
// To be used with Context(...)
func WithSecret(name string, value any) func(context.Context) error {
	return func(ctx context.Context) error {
		cm := cf.SecretsFromContext(ctx)
		return cm.Set(ctx, cf.Ref{Module: optional.Some(ftl.Module()), Name: name}, value)
	}
}

// WithDSN sets a DSN for the current module
//
// To be used with Context(...)
func WithDSN(name string, dbType DBType, dsn string) func(context.Context) error {
	return func(ctx context.Context) error {
		dbProvider := modulecontext.DBProviderFromContext(ctx)
		return dbProvider.Add(name, modulecontext.DBType(dbType), dsn)
	}
}

// WithFakeVerb sets up a mock implementation for a verb
//
// To be used with Context(...)
func WithFakeVerb[Req any, Resp any](verb ftl.Verb[Req, Resp], fake func(ctx context.Context, req Req) (resp Resp, err error)) func(context.Context) error {
	return func(ctx context.Context) error {
		ref := ftl.CallToRef(verb)
		overrider, ok := ftl.CallOverriderFromContext(ctx)
		if !ok {
			return fmt.Errorf("could not override %v with a fake, context not set up with call overrider", ref)
		}
		mockProvider, ok := overrider.(*mockProvider)
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
