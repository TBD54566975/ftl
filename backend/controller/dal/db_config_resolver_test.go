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

func TestDBConfigResolverList(t *testing.T) {
	expected := []configuration.Entry{
		{
			configuration.Ref{
				Module: optional.Some("echo"),
				Name:   "a",
			},
			&url.URL{Scheme: "db"},
		},
		{
			configuration.Ref{
				Module: optional.Some("echo"),
				Name:   "b",
			},
			&url.URL{Scheme: "db"},
		},
		{
			configuration.Ref{
				Module: optional.None[string](),
				Name:   "c",
			},
			&url.URL{Scheme: "db"},
		},
	}

	ctx, resolver, provider := setupDBConfigInterfaces(t)

	// Insert entries in a separate order from how they should be returned to test
	// sorting logic in the SQL query
	_, err := provider.Store(ctx, expected[1].Ref, []byte(`""`))
	assert.NoError(t, err)
	_, err = provider.Store(ctx, expected[2].Ref, []byte(`""`))
	assert.NoError(t, err)
	_, err = provider.Store(ctx, expected[0].Ref, []byte(`""`))
	assert.NoError(t, err)

	entries, err := resolver.List(ctx)
	assert.Equal(t, entries, expected)
	assert.NoError(t, err)
}

func setupDBConfigInterfaces(t *testing.T) (context.Context, DBConfigResolver, DBConfigProvider) {
	t.Helper()

	ctx := log.ContextWithNewDefaultLogger(context.Background())
	conn := sqltest.OpenForTesting(ctx, t)
	dal, err := New(ctx, conn)
	assert.NoError(t, err)
	assert.NotZero(t, dal)

	return ctx, dal.NewConfigResolver(), dal.NewConfigProvider()
}
