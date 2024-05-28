// Package ftltest contains test utilities for the ftl package.
package ftltest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver

	"github.com/TBD54566975/ftl/backend/schema"
	cf "github.com/TBD54566975/ftl/common/configuration"
	pc "github.com/TBD54566975/ftl/common/projectconfig"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/go-runtime/internal"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/modulecontext"
	"github.com/TBD54566975/ftl/internal/slices"
)

type OptionsState struct {
	configs                 map[string][]byte
	secrets                 map[string][]byte
	databases               map[string]modulecontext.Database
	mockVerbs               map[schema.RefKey]modulecontext.Verb
	allowDirectVerbBehavior bool
}

type Option func(context.Context, *OptionsState) error

// Context suitable for use in testing FTL verbs with provided options
func Context(options ...Option) context.Context {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx = internal.WithContext(ctx, newFakeFTL())
	name := reflection.Module()

	state := &OptionsState{
		configs:   make(map[string][]byte),
		secrets:   make(map[string][]byte),
		databases: make(map[string]modulecontext.Database),
		mockVerbs: make(map[schema.RefKey]modulecontext.Verb),
	}
	for _, option := range options {
		err := option(ctx, state)
		if err != nil {
			panic(fmt.Sprintf("error applying option: %v", err))
		}
	}

	builder := modulecontext.NewBuilder(name).AddConfigs(state.configs).AddSecrets(state.secrets).AddDatabases(state.databases)
	builder = builder.UpdateForTesting(state.mockVerbs, state.allowDirectVerbBehavior, newFakeLeaseClient())
	return builder.Build().ApplyToContext(ctx)
}

// WithProjectFiles loads config and secrets from a project file
//
// Takes a list of paths to project files. If multiple paths are provided, they are loaded in order, with later files taking precedence.
// If no paths are provided, the list is inferred from the FTL_CONFIG environment variable. If that is not found, the ftl-project.toml
// file in the git root is used. If ftl-project.toml is not found, no project files are loaded.
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WithProjectFiles("path/to/ftl-project.yaml"),
//		// ... other options
//	)
func WithProjectFiles(paths ...string) Option {
	// Convert to absolute path immediately in case working directory changes
	var preprocessingErr error
	if len(paths) == 0 {
		envValue, ok := os.LookupEnv("FTL_CONFIG")
		if ok {
			paths = strings.Split(envValue, ",")
		} else {
			paths = pc.ConfigPaths(paths)
		}
	}
	paths = slices.Map(paths, func(p string) string {
		path, err := filepath.Abs(p)
		if err != nil {
			preprocessingErr = err
			return ""
		}
		return path
	})
	return func(ctx context.Context, state *OptionsState) error {
		if preprocessingErr != nil {
			return preprocessingErr
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err != nil {
				return fmt.Errorf("error accessing project file: %w", err)
			}
		}
		cm, err := cf.NewDefaultConfigurationManagerFromConfig(ctx, paths)
		if err != nil {
			return fmt.Errorf("could not set up configs: %w", err)
		}
		configs, err := cm.MapForModule(ctx, reflection.Module())
		if err != nil {
			return fmt.Errorf("could not read configs: %w", err)
		}
		for name, data := range configs {
			state.configs[name] = data
		}

		sm, err := cf.NewDefaultSecretsManagerFromConfig(ctx, paths)
		if err != nil {
			return fmt.Errorf("could not set up secrets: %w", err)
		}
		secrets, err := sm.MapForModule(ctx, reflection.Module())
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
//
//	ctx := ftltest.Context(
//		ftltest.WithConfig(exampleEndpoint, "https://example.com"),
//		// ... other options
//	)
func WithConfig[T ftl.ConfigType](config ftl.ConfigValue[T], value T) Option {
	return func(ctx context.Context, state *OptionsState) error {
		if config.Module != reflection.Module() {
			return fmt.Errorf("config %v does not match current module %s", config.Module, reflection.Module())
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
//
//	ctx := ftltest.Context(
//		ftltest.WithSecret(privateKey, "abc123"),
//		// ... other options
//	)
func WithSecret[T ftl.SecretType](secret ftl.SecretValue[T], value T) Option {
	return func(ctx context.Context, state *OptionsState) error {
		if secret.Module != reflection.Module() {
			return fmt.Errorf("secret %v does not match current module %s", secret.Module, reflection.Module())
		}
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		state.secrets[secret.Name] = data
		return nil
	}
}

// WithDatabase sets up a database for testing by appending "_test" to the DSN and emptying all tables
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WithDatabase(db),
//		// ... other options
//	)
func WithDatabase(dbHandle ftl.Database) Option {
	return func(ctx context.Context, state *OptionsState) error {
		originalDSN, err := modulecontext.GetDSNFromSecret(reflection.Module(), dbHandle.Name, state.secrets)
		if err != nil {
			return err
		}

		// convert DSN by appending "_test" to table name
		// postgres DSN format: postgresql://[user[:password]@][netloc][:port][/dbname][?param1=value1&...]
		// source: https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING
		dsnURL, err := url.Parse(originalDSN)
		if err != nil {
			return fmt.Errorf("could not parse DSN: %w", err)
		}
		if dsnURL.Path == "" {
			return fmt.Errorf("DSN for %s must include table name: %s", dbHandle.Name, originalDSN)
		}
		dsnURL.Path += "_test"
		dsn := dsnURL.String()

		// connect to db and clear out the contents of each table
		sqlDB, err := sql.Open("pgx", dsn)
		if err != nil {
			return fmt.Errorf("could not create database %q with DSN %q: %w", dbHandle.Name, dsn, err)
		}
		_, err = sqlDB.ExecContext(ctx, `DO $$
		DECLARE
		   table_name text;
		BEGIN
		   FOR table_name IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public')
		   LOOP
			  EXECUTE 'DELETE FROM ' || table_name;
		   END LOOP;
		END $$;`)
		if err != nil {
			return fmt.Errorf("could not clear tables in database %q: %w", dbHandle.Name, err)
		}

		// replace original database with test database
		replacementDB, err := modulecontext.NewTestDatabase(modulecontext.DBTypePostgres, dsn)
		if err != nil {
			return fmt.Errorf("could not create database %q with DSN %q: %w", dbHandle.Name, dsn, err)
		}
		state.databases[dbHandle.Name] = replacementDB
		return nil
	}
}

// WhenVerb replaces an implementation for a verb
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WhenVerb(Example.Verb, func(ctx context.Context, req Example.Req) (Example.Resp, error) {
//	    	// ...
//		}),
//		// ... other options
//	)
func WhenVerb[Req any, Resp any](verb ftl.Verb[Req, Resp], fake ftl.Verb[Req, Resp]) Option {
	return func(ctx context.Context, state *OptionsState) error {
		ref := reflection.FuncRef(verb)
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

// WhenSource replaces an implementation for a verb with no request
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WhenSource(example.Source, func(ctx context.Context) (example.Resp, error) {
//	    	// ...
//		}),
//		// ... other options
//	)
func WhenSource[Resp any](source ftl.Source[Resp], fake func(ctx context.Context) (resp Resp, err error)) Option {
	return func(ctx context.Context, state *OptionsState) error {
		ref := reflection.FuncRef(source)
		state.mockVerbs[schema.RefKey(ref)] = func(ctx context.Context, req any) (resp any, err error) {
			return fake(ctx)
		}
		return nil
	}
}

// WhenSink replaces an implementation for a verb with no response
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WhenSink(example.Sink, func(ctx context.Context, req example.Req) error {
//	    	...
//		}),
//		// ... other options
//	)
func WhenSink[Req any](sink ftl.Sink[Req], fake func(ctx context.Context, req Req) error) Option {
	return func(ctx context.Context, state *OptionsState) error {
		ref := reflection.FuncRef(sink)
		state.mockVerbs[schema.RefKey(ref)] = func(ctx context.Context, req any) (resp any, err error) {
			request, ok := req.(Req)
			if !ok {
				return nil, fmt.Errorf("invalid request type %T for %v, expected %v", req, ref, reflect.TypeFor[Req]())
			}
			return ftl.Unit{}, fake(ctx, request)
		}
		return nil
	}
}

// WhenEmpty replaces an implementation for a verb with no request or response
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WhenEmpty(Example.Empty, func(ctx context.Context) error {
//	    	...
//		}),
//	)
func WhenEmpty(empty ftl.Empty, fake func(ctx context.Context) (err error)) Option {
	return func(ctx context.Context, state *OptionsState) error {
		ref := reflection.FuncRef(empty)
		state.mockVerbs[schema.RefKey(ref)] = func(ctx context.Context, req any) (resp any, err error) {
			return ftl.Unit{}, fake(ctx)
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
