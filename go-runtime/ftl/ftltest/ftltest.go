// Package ftltest contains test utilities for the ftl package.
package ftltest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/TBD54566975/ftl/backend/schema"
	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/modulecontext"
)

type OptionsState struct {
	configs                 map[string][]byte
	secrets                 map[string][]byte
	databases               map[string]modulecontext.Database
	mockVerbs               map[schema.RefKey]modulecontext.MockVerb
	allowDirectVerbBehavior bool
}

type Option func(context.Context, *OptionsState) error

// Context suitable for use in testing FTL verbs with provided options
func Context(options ...Option) context.Context {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	name := ftl.Module()

	databases, err := modulecontext.DatabasesFromEnvironment(ctx, name)
	if err != nil {
		panic(fmt.Sprintf("error setting up module context from environment: %v", err))
	}

	state := &OptionsState{
		configs:   make(map[string][]byte),
		secrets:   make(map[string][]byte),
		databases: databases,
		mockVerbs: make(map[schema.RefKey]modulecontext.MockVerb),
	}
	for _, option := range options {
		err := option(ctx, state)
		if err != nil {
			panic(fmt.Sprintf("error applying option: %v", err))
		}
	}

	builder := modulecontext.NewBuilder(name).AddConfigs(state.configs).AddSecrets(state.secrets).AddDatabases(map[string]modulecontext.Database{})
	builder = builder.UpdateForTesting(state.mockVerbs, state.allowDirectVerbBehavior)
	return builder.Build().ApplyToContext(ctx)
}

// WithProjectFile loads config and secrets from a project file
//
// To be used when setting up a context for a test:
// ctx := ftltest.Context(
//
//	ftltest.WithProjectFile("path/to/ftl-project.yaml"),
//	... other options
//
// )
func WithProjectFile(path string) Option {
	return func(ctx context.Context, state *OptionsState) error {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("error accessing project file: %w", err)
		}
		cm, err := cf.NewDefaultConfigurationManagerFromConfig(ctx, []string{path})
		if err != nil {
			return fmt.Errorf("could not set up configs: %w", err)
		}
		configs, err := cm.MapForModule(ctx, ftl.Module())
		if err != nil {
			return fmt.Errorf("could not read configs: %w", err)
		}
		for name, data := range configs {
			state.configs[name] = data
		}

		sm, err := cf.NewDefaultSecretsManagerFromConfig(ctx, []string{path})
		if err != nil {
			return fmt.Errorf("could not set up secrets: %w", err)
		}
		secrets, err := sm.MapForModule(ctx, ftl.Module())
		if err != nil {
			return fmt.Errorf("could not read secrets: %w", err)
		}
		for name, data := range secrets {
			state.secrets[name] = data
		}
		return nil
	}
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
func WithConfig[T ftl.ConfigType](config ftl.ConfigValue[T], value T) Option {
	return func(ctx context.Context, state *OptionsState) error {
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
func WithSecret[T ftl.SecretType](secret ftl.SecretValue[T], value T) Option {
	return func(ctx context.Context, state *OptionsState) error {
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
func WhenVerb[Req any, Resp any](verb ftl.Verb[Req, Resp], fake func(ctx context.Context, req Req) (resp Resp, err error)) Option {
	return func(ctx context.Context, state *OptionsState) error {
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
func WithCallsAllowedWithinModule() Option {
	return func(ctx context.Context, state *OptionsState) error {
		state.allowDirectVerbBehavior = true
		return nil
	}
}
