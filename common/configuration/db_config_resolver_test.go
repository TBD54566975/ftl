package configuration

import (
	"context"
	"testing"

	"github.com/TBD54566975/ftl/common/configuration/sql"
	"github.com/alecthomas/assert/v2"
)

type mockDBResolverDAL struct{}

func (mockDBResolverDAL) ListModuleConfiguration(ctx context.Context) ([]sql.ModuleConfiguration, error) {
	return []sql.ModuleConfiguration{}, nil
}

func (mockDBResolverDAL) ListModuleSecrets(ctx context.Context) ([]sql.ModuleSecret, error) {
	return []sql.ModuleSecret{}, nil
}

func TestDBConfigResolverList(t *testing.T) {
	ctx := context.Background()
	resolver := NewDBResolver[Configuration](mockDBResolverDAL{})
	expected := []Entry{}

	entries, err := resolver.List(ctx)
	assert.Equal(t, entries, expected)
	assert.NoError(t, err)
}
