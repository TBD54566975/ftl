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

func TestDBResolverRoundTrip(t *testing.T) {
	tests := []struct {
		TestName     string
		ModuleSet    optional.Option[string]
		ModuleGet    optional.Option[string]
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

	ctx, cr := setup(t)
	u := URL("db://asdfasdf")
	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			if test.PresetGlobal {
				err := cr.Set(ctx, configuration.Ref{
					Module: optional.None[string](),
					Name:   "configname",
				}, URL("db://qwerty"))
				assert.NoError(t, err)
			}
			err := cr.Set(ctx, configuration.Ref{
				Module: test.ModuleSet,
				Name:   "configname",
			}, u)
			assert.NoError(t, err)
			gotURL, err := cr.Get(ctx, configuration.Ref{
				Module: test.ModuleGet,
				Name:   "configname",
			})
			assert.NoError(t, err)
			assert.Equal(t, u, gotURL)
			err = cr.Unset(ctx, configuration.Ref{
				Module: test.ModuleSet,
				Name:   "configname",
			})
			assert.NoError(t, err)
		})
	}
}

func setup(t *testing.T) (context.Context, DatabaseResolver) {
	t.Helper()

	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	dal, err := New(ctx, conn)
	assert.NoError(t, err)
	assert.NotZero(t, dal)

	return ctx, dal.NewConfigResolver()
}

func URL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}
