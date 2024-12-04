// Package ftltest contains test utilities for the ftl package.
package ftltest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib" // SQL driver

	"github.com/TBD54566975/ftl/backend/provisioner"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/go-runtime/internal"
	"github.com/TBD54566975/ftl/go-runtime/server"
	cf "github.com/TBD54566975/ftl/internal/configuration/manager"
	"github.com/TBD54566975/ftl/internal/configuration/providers"
	"github.com/TBD54566975/ftl/internal/deploymentcontext"
	"github.com/TBD54566975/ftl/internal/log"
	pc "github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/schema"
	mcu "github.com/TBD54566975/ftl/internal/testutils/modulecontext"
)

// Allows tests to mock module reflection
var moduleGetter = reflection.Module

type OptionsState struct {
	databases               map[string]deploymentcontext.Database
	mockVerbs               map[schema.RefKey]deploymentcontext.Verb
	allowDirectVerbBehavior bool
}

type optionRank int

const (
	profile optionRank = iota
	other
)

type Option struct {
	rank  optionRank
	apply func(context.Context, *OptionsState) error
}

// Context suitable for use in testing FTL verbs with provided options
func Context(options ...Option) context.Context {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	module := moduleGetter()
	return newContext(ctx, module, options...)
}

func newContext(ctx context.Context, module string, options ...Option) context.Context {
	state := &OptionsState{
		databases: make(map[string]deploymentcontext.Database),
		mockVerbs: make(map[schema.RefKey]deploymentcontext.Verb),
	}

	ctx = contextWithFakeFTL(ctx, options...)

	sort.Slice(options, func(i, j int) bool {
		return options[i].rank < options[j].rank
	})

	for _, option := range options {
		err := option.apply(ctx, state)
		if err != nil {
			panic(fmt.Sprintf("error applying option: %v", err))
		}
	}

	builder := deploymentcontext.NewBuilder(module).AddDatabases(state.databases)
	builder = builder.UpdateForTesting(state.mockVerbs, state.allowDirectVerbBehavior, newFakeLeaseClient())

	return mcu.MakeDynamic(ctx, builder.Build()).ApplyToContext(ctx)
}

// SubContext applies the given options to the given context, creating a new
// context extending the previous one.
//
// Does not modify the existing context
func SubContext(ctx context.Context, options ...Option) context.Context {
	oldFtl := internal.FromContext(ctx).(*fakeFTL) //nolint:forcetypeassert
	module := moduleGetter()
	return newContext(ctx, module, append(oldFtl.options, options...)...)
}

// WithDefaultProjectFile loads config and secrets from the default project
// file, which is either the FTL_CONFIG environment variable or the
// ftl-project.toml file in the git root.
func WithDefaultProjectFile() Option {
	return WithProjectFile("")
}

// WithProjectFile loads config and secrets from a project file
//
// Takes a path to an FTL project file. If an empty path is provided, the path
// is inferred from the FTL_CONFIG environment variable. If that is not found,
// the ftl-project.toml file in the git root is used. If a project file is not
// found, an error is returned.
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WithProjectFile("path/to/ftl-project.yaml"),
//		// ... other options
//	)
func WithProjectFile(path string) Option {
	// Convert to absolute path immediately in case working directory changes
	var preprocessingErr error
	if path == "" {
		var ok bool
		path, ok = pc.DefaultConfigPath().Get()
		if !ok {
			preprocessingErr = fmt.Errorf("could not find default project file in $FTL_CONFIG or git")
		}
	}
	return Option{
		rank: profile,
		apply: func(ctx context.Context, state *OptionsState) error {
			if preprocessingErr != nil {
				return preprocessingErr
			}
			if _, err := os.Stat(path); err != nil {
				return fmt.Errorf("error accessing project file: %w", err)
			}
			projectConfig, err := pc.Load(ctx, path)
			if err != nil {
				return fmt.Errorf("project: %w", err)
			}
			cm, err := cf.NewDefaultConfigurationManagerFromConfig(ctx, providers.NewDefaultConfigRegistry(), projectConfig)
			if err != nil {
				return fmt.Errorf("could not set up configs: %w", err)
			}
			configs, err := cm.MapForModule(ctx, moduleGetter())
			if err != nil {
				return fmt.Errorf("could not read configs: %w", err)
			}

			fftl := internal.FromContext(ctx).(*fakeFTL) //nolint:forcetypeassert
			for name, data := range configs {
				if err := fftl.setConfig(name, json.RawMessage(data)); err != nil {
					return err
				}
			}

			sm, err := cf.NewDefaultSecretsManagerFromConfig(ctx, providers.NewDefaultSecretsRegistry(projectConfig, ""), projectConfig)
			if err != nil {
				return fmt.Errorf("could not set up secrets: %w", err)
			}
			secrets, err := sm.MapForModule(ctx, moduleGetter())
			if err != nil {
				return fmt.Errorf("could not read secrets: %w", err)
			}
			for name, data := range secrets {
				if err := fftl.setSecret(name, json.RawMessage(data)); err != nil {
					return err
				}
			}
			return nil
		},
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
func WithConfig[T ftl.ConfigType](config ftl.Config[T], value T) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			if config.Module != moduleGetter() {
				return fmt.Errorf("config %v does not match current module %s", config.Module, moduleGetter())
			}
			fftl := internal.FromContext(ctx).(*fakeFTL) //nolint:forcetypeassert
			if err := fftl.setConfig(config.Name, value); err != nil {
				return err
			}
			return nil
		},
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
func WithSecret[T ftl.SecretType](secret ftl.Secret[T], value T) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			if secret.Module != moduleGetter() {
				return fmt.Errorf("secret %v does not match current module %s", secret.Module, moduleGetter())
			}
			fftl := internal.FromContext(ctx).(*fakeFTL) //nolint:forcetypeassert
			if err := fftl.setSecret(secret.Name, value); err != nil {
				return err
			}
			return nil
		},
	}
}

// WithDatabase sets up a database for testing by appending "_test" to the DSN and emptying all tables
func WithDatabase[T ftl.DatabaseConfig]() Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			cfg := defaultDatabaseConfig[T]()
			name := cfg.Name()
			switch any(cfg).(type) {
			case ftl.PostgresDatabaseConfig:
				dsn, err := provisioner.ProvisionPostgresForTest(ctx, moduleGetter(), name)
				if err != nil {
					return fmt.Errorf("could not provision database %q: %w", name, err)
				}
				dir, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("could not get working dir")
				}
				err = provisioner.RunPostgresMigration(ctx, dsn, dir, name)
				if err != nil {
					return fmt.Errorf("could not migrate database %q: %w", name, err)
				}
				// replace original database with test database
				replacementDB, err := deploymentcontext.NewTestDatabase(deploymentcontext.DBTypePostgres, dsn)
				if err != nil {
					return fmt.Errorf("could not create database %q with DSN %q: %w", name, dsn, err)
				}
				state.databases[name] = replacementDB
			case ftl.MySQLDatabaseConfig:
				dsn, err := provisioner.ProvisionMySQLForTest(ctx, moduleGetter(), name)
				if err != nil {
					return fmt.Errorf("could not provision database %q: %w", name, err)
				}
				dir, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("could not get working dir")
				}
				err = provisioner.RunMySQLMigration(ctx, dsn, dir, name)
				if err != nil {
					return fmt.Errorf("could not migrate database %q: %w", name, err)
				}
				// replace original database with test database
				replacementDB, err := deploymentcontext.NewTestDatabase(deploymentcontext.DBTypeMySQL, dsn)
				if err != nil {
					return fmt.Errorf("could not create database %q with DSN %q: %w", name, dsn, err)
				}
				state.databases[name] = replacementDB

			}
			return nil
		},
	}
}

// WhenVerb replaces an implementation for a verb
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WhenVerb[example.VerbClient](func(ctx context.Context, req example.Req) (example.Resp, error) {
//	    	// ...
//		}),
//		// ... other options
//	)
func WhenVerb[VerbClient, Req, Resp any](fake ftl.Verb[Req, Resp]) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			ref := reflection.ClientRef[VerbClient]()
			state.mockVerbs[schema.RefKey(ref)] = func(ctx context.Context, req any) (resp any, err error) {
				request, ok := req.(Req)
				if !ok {
					return nil, fmt.Errorf("invalid request type %T for %v, expected %v", req, ref, reflect.TypeFor[Req]())
				}
				return fake(ctx, request)
			}
			return nil
		},
	}
}

// WhenSource replaces an implementation for a verb with no request
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WhenSource[example.SourceClient](func(ctx context.Context) (example.Resp, error) {
//	    	// ...
//		}),
//		// ... other options
//	)
func WhenSource[SourceClient, Resp any](fake ftl.Source[Resp]) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			ref := reflection.ClientRef[SourceClient]()
			state.mockVerbs[schema.RefKey(ref)] = func(ctx context.Context, req any) (resp any, err error) {
				return fake(ctx)
			}
			return nil
		},
	}
}

// WhenSink replaces an implementation for a verb with no response
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WhenSink[example.SinkClient](func(ctx context.Context, req example.Req) error {
//	    	...
//		}),
//		// ... other options
//	)
func WhenSink[SinkClient, Req any](fake ftl.Sink[Req]) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			ref := reflection.ClientRef[SinkClient]()
			state.mockVerbs[schema.RefKey(ref)] = func(ctx context.Context, req any) (resp any, err error) {
				request, ok := req.(Req)
				if !ok {
					return nil, fmt.Errorf("invalid request type %T for %v, expected %v", req, ref, reflect.TypeFor[Req]())
				}
				return ftl.Unit{}, fake(ctx, request)
			}
			return nil
		},
	}
}

// WhenEmpty replaces an implementation for a verb with no request or response
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WhenEmpty[example.EmptyClient](func(ctx context.Context) error {
//	    	...
//		}),
//	)
func WhenEmpty[EmptyClient any](fake ftl.Empty) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			ref := reflection.ClientRef[EmptyClient]()
			state.mockVerbs[schema.RefKey(ref)] = func(ctx context.Context, req any) (resp any, err error) {
				return ftl.Unit{}, fake(ctx)
			}
			return nil
		},
	}
}

// WithCallsAllowedWithinModule allows tests to enable calls to all verbs within the current module
//
// Any overrides provided by calling WhenVerb(...) will take precedence
func WithCallsAllowedWithinModule() Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			state.allowDirectVerbBehavior = true
			return nil
		},
	}
}

// WhenMap injects a fake implementation of a Mapping function
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WhenMap(Example.MapHandle, func(ctx context.Context) (U, error) {
//	    	// ...
//		}),
//		// ... other options
//	)
func WhenMap[T, U any](mapper *ftl.MapHandle[T, U], fake func(context.Context) (U, error)) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			fftl := internal.FromContext(ctx).(*fakeFTL) //nolint:forcetypeassert
			addMapMock(fftl, mapper, fake)
			return nil
		},
	}
}

// WithMapsAllowed allows all `ftl.Map` calls to pass through to their original
// implementation.
//
// Any overrides provided by calling WhenMap(...) will take precedence.
func WithMapsAllowed() Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			fftl := internal.FromContext(ctx).(*fakeFTL) //nolint:forcetypeassert
			fftl.startAllowingMapCalls()
			return nil
		},
	}
}

// dsnSecretKey returns the key for the secret that is expected to hold the DSN for a database.
//
// The format is FTL_DSN_<MODULE>_<DBNAME>
func dsnSecretKey(module, name string) string {
	return fmt.Sprintf("FTL_DSN_%s_%s", strings.ToUpper(module), strings.ToUpper(name))
}

// getDSNFromSecret returns the DSN for a database from the relevant secret
func getDSNFromSecret(ftl internal.FTL, module, name string) (string, error) {
	key := dsnSecretKey(module, name)
	var dsn string
	if err := ftl.GetSecret(context.Background(), key, &dsn); err != nil {
		return "", fmt.Errorf("could not get DSN for database %q from secret %q: %w", name, key, err)
	}
	return dsn, nil
}

// Call a Verb inline, applying resources and test behavior.
func Call[VerbClient, Req, Resp any](ctx context.Context, req Req) (Resp, error) {
	return call[VerbClient, Req, Resp](ctx, req)
}

// CallSource calls a Source inline, applying resources and test behavior.
func CallSource[VerbClient, Resp any](ctx context.Context) (Resp, error) {
	return call[VerbClient, ftl.Unit, Resp](ctx, ftl.Unit{})
}

// CallSink calls a Sink inline, applying resources and test behavior.
func CallSink[VerbClient, Req any](ctx context.Context, req Req) error {
	_, err := call[VerbClient, Req, ftl.Unit](ctx, req)
	return err
}

// CallEmpty calls an Empty inline, applying resources and test behavior.
func CallEmpty[VerbClient any](ctx context.Context) error {
	_, err := call[VerbClient, ftl.Unit, ftl.Unit](ctx, ftl.Unit{})
	return err
}

// GetDatabaseHandle returns a database handle using the given database config.
func GetDatabaseHandle[T ftl.DatabaseConfig]() (ftl.DatabaseHandle[T], error) {
	reflectedDB := reflection.GetDatabase[T]()
	if reflectedDB == nil {
		return ftl.DatabaseHandle[T]{}, fmt.Errorf("could not find database for config")
	}

	var dbType ftl.DatabaseType
	switch reflectedDB.DBType {
	case "postgres":
		dbType = ftl.DatabaseTypePostgres
	case "mysql":
		dbType = ftl.DatabaseTypeMysql
	default:
		return ftl.DatabaseHandle[T]{}, fmt.Errorf("unsupported database type %v", reflectedDB.DBType)
	}
	return ftl.NewDatabaseHandle[T](defaultDatabaseConfig[T](), dbType, reflectedDB.DB), nil
}

func call[VerbClient, Req, Resp any](ctx context.Context, req Req) (resp Resp, err error) {
	ref := reflection.ClientRef[VerbClient]()
	// always allow direct behavior for the verb triggered by this call
	moduleCtx := deploymentcontext.NewBuilderFromContext(
		deploymentcontext.FromContext(ctx).CurrentContext(),
	).AddAllowedDirectVerb(ref).Build()
	ctx = mcu.MakeDynamic(ctx, moduleCtx).ApplyToContext(ctx)

	inline := server.InvokeVerb[Req, Resp](ref)
	override, err := moduleCtx.BehaviorForVerb(schema.Ref{Module: ref.Module, Name: ref.Name})
	if err != nil {
		return resp, fmt.Errorf("test harness failed to retrieve behavior for verb %s: %w", ref, err)
	}
	if behavior, ok := override.Get(); ok {
		uncheckedResp, err := behavior.Call(ctx, deploymentcontext.Verb(widenVerb(inline)), req)
		if err != nil {
			return resp, fmt.Errorf("test harness failed to call verb %s: %w", ref, err)
		}
		if r, ok := uncheckedResp.(Resp); ok {
			return r, nil
		}
		return resp, fmt.Errorf("%s: overridden verb had invalid response type %T, expected %v", ref, uncheckedResp, reflect.TypeFor[Resp]())
	}
	return inline(ctx, req)
}

func widenVerb[Req, Resp any](verb ftl.Verb[Req, Resp]) ftl.Verb[any, any] {
	return func(ctx context.Context, uncheckedReq any) (any, error) {
		req, ok := uncheckedReq.(Req)
		if !ok {
			return nil, fmt.Errorf("invalid request type %T for %v, expected %v", uncheckedReq, reflection.FuncRef(verb), reflect.TypeFor[Req]())
		}
		return verb(ctx, req)
	}
}

func defaultDatabaseConfig[T ftl.DatabaseConfig]() T {
	typ := reflect.TypeFor[T]()
	var cfg T
	if typ.Kind() == reflect.Ptr {
		cfg = reflect.New(typ.Elem()).Interface().(T) //nolint:forcetypeassert
	} else {
		cfg = reflect.New(typ).Elem().Interface().(T) //nolint:forcetypeassert
	}
	return cfg
}
