package providers

import (
	"context"
	"net/url"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/configuration"
)

var b = []byte(`""`)

type mockDBConfigProviderDAL struct{}

func (mockDBConfigProviderDAL) GetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) ([]byte, error) {
	return b, nil
}

func (mockDBConfigProviderDAL) SetModuleConfiguration(ctx context.Context, module optional.Option[string], name string, value []byte) error {
	return nil
}

func (mockDBConfigProviderDAL) UnsetModuleConfiguration(ctx context.Context, module optional.Option[string], name string) error {
	return nil
}

func TestDBConfigProvider(t *testing.T) {
	ctx := context.Background()
	provider := NewDatabaseConfig(mockDBConfigProviderDAL{})

	gotBytes, err := provider.Load(ctx, configuration.Ref{
		Module: optional.Some("module"),
		Name:   "configname",
	}, &url.URL{Scheme: "db"})
	assert.NoError(t, err)
	assert.Equal(t, b, gotBytes)

	gotURL, err := provider.Store(ctx, configuration.Ref{
		Module: optional.Some("module"),
		Name:   "configname",
	}, b)
	assert.NoError(t, err)
	assert.Equal(t, &url.URL{Scheme: "db"}, gotURL)

	err = provider.Delete(ctx, configuration.Ref{
		Module: optional.Some("module"),
		Name:   "configname",
	})
	assert.NoError(t, err)
}

func TestDBConfigProvider_Global(t *testing.T) {
	t.Run("works", func(t *testing.T) {
		ctx := context.Background()
		provider := NewDatabaseConfig(mockDBConfigProviderDAL{})

		gotBytes, err := provider.Load(ctx, configuration.Ref{
			Module: optional.None[string](),
			Name:   "configname",
		}, &url.URL{Scheme: "db"})
		assert.NoError(t, err)
		assert.Equal(t, b, gotBytes)

		gotURL, err := provider.Store(ctx, configuration.Ref{
			Module: optional.None[string](),
			Name:   "configname",
		}, b)
		assert.NoError(t, err)
		assert.Equal(t, &url.URL{Scheme: "db"}, gotURL)

		err = provider.Delete(ctx, configuration.Ref{
			Module: optional.None[string](),
			Name:   "configname",
		})
		assert.NoError(t, err)
	})

	// TODO: maybe add a unit test to assert failure to create same global config twice. not sure how to wire up the mocks for this
}
