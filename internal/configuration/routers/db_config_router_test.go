package routers

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/configuration/dal"
)

type mockDBConfigResolverDAL struct{}

func (mockDBConfigResolverDAL) ListModuleConfiguration(ctx context.Context) ([]dal.ModuleConfiguration, error) {
	return []dal.ModuleConfiguration{}, nil
}

func TestDBConfigResolverList(t *testing.T) {
	ctx := context.Background()
	resolver := NewDatabaseConfig(mockDBConfigResolverDAL{})
	expected := []configuration.Entry{}

	entries, err := resolver.List(ctx)
	assert.Equal(t, entries, expected)
	assert.NoError(t, err)
}
