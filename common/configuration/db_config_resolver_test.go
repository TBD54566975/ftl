package configuration

import (
	"context"
	"testing"

	"github.com/TBD54566975/ftl/common/configuration/sql"
	"github.com/alecthomas/assert/v2"
)

type mockDBConfigResolverDAL struct{}

func (mockDBConfigResolverDAL) ListModuleConfiguration(ctx context.Context) ([]sql.ModuleConfiguration, error) {
	return []sql.ModuleConfiguration{}, nil
}

func TestDBConfigResolverList(t *testing.T) {
	ctx := context.Background()
	resolver := NewDBConfigResolver(mockDBConfigResolverDAL{})
	expected := []Entry{}

	entries, err := resolver.List(ctx)
	assert.Equal(t, entries, expected)
	assert.NoError(t, err)
}
