// Package ftltest contains test utilities for the ftl package.
package ftltest

import (
	"context"
	"fmt"
	"reflect"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/modulecontext"
	"github.com/TBD54566975/ftl/internal/log"
)

// Context suitable for use in testing FTL verbs with provided options
func Context(options ...func(context.Context) error) context.Context {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	context, err := modulecontext.FromEnvironment(ctx, ftl.Module(), true)
	if err != nil {
		panic(err)
	}
	ctx = context.ApplyToContext(ctx)

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
		if config.Module != ftl.Module() {
			return fmt.Errorf("config %v does not match current module %s", config.Module, ftl.Module())
		}
		return modulecontext.FromContext(ctx).SetConfig(config.Name, value)
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
		if secret.Module != ftl.Module() {
			return fmt.Errorf("secret %v does not match current module %s", secret.Module, ftl.Module())
		}
		return modulecontext.FromContext(ctx).SetSecret(secret.Name, value)
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
		ref := ftl.FuncRef(verb)
		modulecontext.FromContext(ctx).SetMockVerb(modulecontext.Ref(ref), func(ctx context.Context, req any) (resp any, err error) {
			request, ok := req.(Req)
			if !ok {
				return nil, fmt.Errorf("invalid request type %T for %v, expected %v", req, ref, reflect.TypeFor[Req]())
			}
			return fake(ctx, request)
		})
		return nil
	}
}

// WithCallsAllowedWithinModule allows tests to enable calls to all verbs within the current module
//
// Any overrides provided by calling WhenVerb(...) will take precedence
func WithCallsAllowedWithinModule() func(context.Context) error {
	return func(ctx context.Context) error {
		modulecontext.FromContext(ctx).AllowDirectVerbBehaviorWithinModule()
		return nil
	}
}
