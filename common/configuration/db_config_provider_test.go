package configuration

import (
	"context"
	"net/url"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
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
	provider := NewDBConfigProvider(mockDBConfigProviderDAL{})

	gotWrapper, err := provider.Load(ctx, Ref{
		Module: optional.Some("module"),
		Name:   "configname",
	}, &url.URL{Scheme: "db"})
	assert.NoError(t, err)
	gotBytes, err := gotWrapper.Unwrap(optional.None[Obfuscator]())
	assert.NoError(t, err)
	assert.Equal(t, b, gotBytes)

	gotURL, err := provider.Store(ctx, Ref{
		Module: optional.Some("module"),
		Name:   "configname",
	}, b)
	assert.NoError(t, err)
	assert.Equal(t, &url.URL{Scheme: "db"}, gotURL)

	err = provider.Delete(ctx, Ref{
		Module: optional.Some("module"),
		Name:   "configname",
	})
	assert.NoError(t, err)
}
