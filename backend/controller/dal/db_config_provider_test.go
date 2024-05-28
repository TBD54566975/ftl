package dal

import (
	"context"
	"net/url"
	"testing"

	"github.com/TBD54566975/ftl/backend/controller/sql/sqltest"
	"github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
)

func TestDBConfigProviderRoundTrip(t *testing.T) {
	tests := []struct {
		TestName     string
		ModuleStore  optional.Option[string]
		ModuleLoad   optional.Option[string]
		PresetGlobal bool
	}{
		{
			"SetModuleGetModule",
			optional.Some("echo"),
			optional.Some("echo"),
			false,
		},
		{
			"SetGlobalGetGlobal",
			optional.None[string](),
			optional.None[string](),
			false,
		},
		{
			"SetGlobalGetModule",
			optional.None[string](),
			optional.Some("echo"),
			false,
		},
		{
			"SetModuleOverridesGlobal",
			optional.Some("echo"),
			optional.Some("echo"),
			true,
		},
	}

	ctx, provider := setupDBConfigProvider(t)
	b := []byte(`"asdf"`)
	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			if test.PresetGlobal {
				_, err := provider.Store(ctx, configuration.Ref{
					Module: optional.None[string](),
					Name:   "configname",
				}, []byte(`"qwerty"`))
				assert.NoError(t, err)
			}
			_, err := provider.Store(ctx, configuration.Ref{
				Module: test.ModuleStore,
				Name:   "configname",
			}, b)
			assert.NoError(t, err)
			gotBytes, err := provider.Load(ctx, configuration.Ref{
				Module: test.ModuleLoad,
				Name:   "configname",
			}, &url.URL{Scheme: "db"})
			assert.NoError(t, err)
			assert.Equal(t, b, gotBytes)
			err = provider.Delete(ctx, configuration.Ref{
				Module: test.ModuleStore,
				Name:   "configname",
			})
			assert.NoError(t, err)
		})
	}
}

func setupDBConfigProvider(t *testing.T) (context.Context, DBConfigProvider) {
	t.Helper()

	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	dal, err := New(ctx, conn)
	assert.NoError(t, err)
	assert.NotZero(t, dal)

	return ctx, dal.NewConfigProvider()
}
