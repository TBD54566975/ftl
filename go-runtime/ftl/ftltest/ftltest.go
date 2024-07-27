// Package ftltest contains test utilities for the ftl package.
package ftltest

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"sort"
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
	mcu "github.com/TBD54566975/ftl/testutils/modulecontext"
)

type OptionsState struct {
	databases               map[string]modulecontext.Database
	mockVerbs               map[schema.RefKey]modulecontext.Verb
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
	state := &OptionsState{
		databases: make(map[string]modulecontext.Database),
		mockVerbs: make(map[schema.RefKey]modulecontext.Verb),
	}

	ctx := log.ContextWithNewDefaultLogger(context.Background())
	ctx, _ = newFakeFTL(ctx)
	name := reflection.Module()

	sort.Slice(options, func(i, j int) bool {
		return options[i].rank < options[j].rank
	})

	for _, option := range options {
		err := option.apply(ctx, state)
		if err != nil {
			panic(fmt.Sprintf("error applying option: %v", err))
		}
	}

	builder := modulecontext.NewBuilder(name).AddDatabases(state.databases)
	builder = builder.UpdateForTesting(state.mockVerbs, state.allowDirectVerbBehavior, newFakeLeaseClient())
	return mcu.MakeDynamic(ctx, builder.Build()).ApplyToContext(ctx)
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
			cm, err := cf.NewDefaultConfigurationManagerFromConfig(ctx, path)
			if err != nil {
				return fmt.Errorf("could not set up configs: %w", err)
			}
			configs, err := cm.MapForModule(ctx, reflection.Module())
			if err != nil {
				return fmt.Errorf("could not read configs: %w", err)
			}

			fftl := internal.FromContext(ctx).(*fakeFTL) //nolint:forcetypeassert
			for name, data := range configs {
				if err := fftl.setConfig(name, json.RawMessage(data)); err != nil {
					return err
				}
			}

			sm, err := cf.NewDefaultSecretsManagerFromConfig(ctx, path, "")
			if err != nil {
				return fmt.Errorf("could not set up secrets: %w", err)
			}
			secrets, err := sm.MapForModule(ctx, reflection.Module())
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
func WithConfig[T ftl.ConfigType](config ftl.ConfigValue[T], value T) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			if config.Module != reflection.Module() {
				return fmt.Errorf("config %v does not match current module %s", config.Module, reflection.Module())
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
func WithSecret[T ftl.SecretType](secret ftl.SecretValue[T], value T) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			if secret.Module != reflection.Module() {
				return fmt.Errorf("secret %v does not match current module %s", secret.Module, reflection.Module())
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
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WithDatabase(db),
//		// ... other options
//	)
func WithDatabase(dbHandle ftl.Database) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			fftl := internal.FromContext(ctx)
			originalDSN, err := getDSNFromSecret(fftl, reflection.Module(), dbHandle.Name)
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
				EXECUTE 'ALTER TABLE ' || quote_ident(table_name) || ' DISABLE TRIGGER ALL;';
				EXECUTE 'DELETE FROM ' || quote_ident(table_name) || ';';
				EXECUTE 'ALTER TABLE ' || quote_ident(table_name) || ' ENABLE TRIGGER ALL;';
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
		},
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
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			ref := reflection.FuncRef(verb)
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
//		ftltest.WhenSource(example.Source, func(ctx context.Context) (example.Resp, error) {
//	    	// ...
//		}),
//		// ... other options
//	)
func WhenSource[Resp any](source ftl.Source[Resp], fake func(ctx context.Context) (resp Resp, err error)) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			ref := reflection.FuncRef(source)
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
//		ftltest.WhenSink(example.Sink, func(ctx context.Context, req example.Req) error {
//	    	...
//		}),
//		// ... other options
//	)
func WhenSink[Req any](sink ftl.Sink[Req], fake func(ctx context.Context, req Req) error) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			ref := reflection.FuncRef(sink)
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
//		ftltest.WhenEmpty(Example.Empty, func(ctx context.Context) error {
//	    	...
//		}),
//	)
func WhenEmpty(empty ftl.Empty, fake func(ctx context.Context) (err error)) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			ref := reflection.FuncRef(empty)
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

// WithSubscriber adds a subscriber during a test
//
// By default, all subscribers are disabled in unit tests, and must be manually enabled by calling WithSubscriber(…).
// This allows easy isolation for each unit test.
//
// WithSubscriber(…) can also be used to make an ad-hoc subscriber for your test by defining a new function as the sink.
//
// To be used when setting up a context for a test:
//
//	ctx := ftltest.Context(
//		ftltest.WithSubscriber(paymentTopic, ProcessPayment),
//		// ... other options
//	)
func WithSubscriber[E any](subscription ftl.SubscriptionHandle[E], sink ftl.Sink[E]) Option {
	return Option{
		rank: other,
		apply: func(ctx context.Context, state *OptionsState) error {
			fftl := internal.FromContext(ctx).(*fakeFTL) //nolint:forcetypeassert
			addSubscriber(fftl.pubSub, subscription, sink)
			return nil
		},
	}
}

// EventsForTopic returns all published events for a topic
func EventsForTopic[E any](ctx context.Context, topic ftl.TopicHandle[E]) []E {
	fftl := internal.FromContext(ctx).(*fakeFTL) //nolint:forcetypeassert
	return eventsForTopic(ctx, fftl.pubSub, topic)
}

type SubscriptionResult[E any] struct {
	Event E
	Error ftl.Option[error]
}

// ResultsForSubscription returns all consumed events for a subscription, with any resulting errors
func ResultsForSubscription[E any](ctx context.Context, subscription ftl.SubscriptionHandle[E]) []SubscriptionResult[E] {
	fftl := internal.FromContext(ctx).(*fakeFTL) //nolint:forcetypeassert
	return resultsForSubscription(ctx, fftl.pubSub, subscription)
}

// ErrorsForSubscription returns all errors encountered while consuming events for a subscription
func ErrorsForSubscription[E any](ctx context.Context, subscription ftl.SubscriptionHandle[E]) []error {
	errs := []error{}
	for _, result := range ResultsForSubscription(ctx, subscription) {
		if err, ok := result.Error.Get(); ok {
			errs = append(errs, err)
		}
	}
	return errs
}

// WaitForSubscriptionsToComplete waits until all subscriptions have consumed all events
//
// Subscriptions with no manually activated subscribers are ignored.
// Make sure you have called WithSubscriber(…) for all subscriptions you want to wait for.
func WaitForSubscriptionsToComplete(ctx context.Context) {
	fftl := internal.FromContext(ctx).(*fakeFTL) //nolint:forcetypeassert
	fftl.pubSub.waitForSubscriptionsToComplete(ctx)
}
