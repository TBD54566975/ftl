// Package ftltest contains test utilities for the ftl package.
package ftltest

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/modulecontext"
	"github.com/TBD54566975/ftl/internal/log"
)

type Options struct {
	configs                 map[string][]byte
	secrets                 map[string][]byte
	mockVerbs               map[schema.RefKey]modulecontext.MockVerb
	allowDirectVerbBehavior bool
}

// Context suitable for use in testing FTL verbs with provided options
func Context(options ...func(*Options) error) context.Context {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	state := &Options{
		configs:   make(map[string][]byte),
		secrets:   make(map[string][]byte),
		mockVerbs: make(map[schema.RefKey]modulecontext.MockVerb),
	}
	for _, option := range options {
		err := option(state)
		if err != nil {
			panic(fmt.Sprintf("error applying option: %v", err))
		}
	}

	builder, err := modulecontext.NewBuilder(ftl.Module()).UpdateFromEnvironment(ctx)
	if err != nil {
		panic(fmt.Sprintf("error setting up module context from environment: %v", err))
	}
	builder = builder.AddConfigs(state.configs).AddSecrets(state.secrets).AddDatabases(map[string]modulecontext.Database{})
	builder = builder.UpdateForTesting(state.mockVerbs, state.allowDirectVerbBehavior)
	return builder.Build().ApplyToContext(ctx)
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
func WithConfig[T ftl.ConfigType](config ftl.ConfigValue[T], value T) func(*Options) error {
	return func(state *Options) error {
		if config.Module != ftl.Module() {
			return fmt.Errorf("config %v does not match current module %s", config.Module, ftl.Module())
		}
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		state.configs[config.Name] = data
		return nil
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
func WithSecret[T ftl.SecretType](secret ftl.SecretValue[T], value T) func(*Options) error {
	return func(state *Options) error {
		if secret.Module != ftl.Module() {
			return fmt.Errorf("secret %v does not match current module %s", secret.Module, ftl.Module())
		}
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		state.secrets[secret.Name] = data
		return nil
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
func WhenVerb[Req any, Resp any](verb ftl.Verb[Req, Resp], fake func(ctx context.Context, req Req) (resp Resp, err error)) func(*Options) error {
	return func(state *Options) error {
		ref := ftl.FuncRef(verb)
		state.mockVerbs[schema.RefKey(ref)] = func(ctx context.Context, req any) (resp any, err error) {
			request, ok := req.(Req)
			if !ok {
				return nil, fmt.Errorf("invalid request type %T for %v, expected %v", req, ref, reflect.TypeFor[Req]())
			}
			return fake(ctx, request)
		}
		return nil
	}
}

// WithCallsAllowedWithinModule allows tests to enable calls to all verbs within the current module
//
// Any overrides provided by calling WhenVerb(...) will take precedence
func WithCallsAllowedWithinModule() func(*Options) error {
	return func(state *Options) error {
		state.allowDirectVerbBehavior = true
		return nil
	}
}
