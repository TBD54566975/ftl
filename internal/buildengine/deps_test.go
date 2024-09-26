package buildengine

import (
	"context"
	"testing"

	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/alecthomas/assert/v2"
)

func TestExtractModuleDepsGo(t *testing.T) {
	ctx := context.Background()
	config, err := moduleconfig.LoadModuleConfig("testdata/alpha")
	assert.NoError(t, err)

	plugin, err := PluginFromConfig(ctx, config.Abs(), "")
	assert.NoError(t, err)

	deps, err := plugin.GetDependencies(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"another", "other"}, deps)
}

func TestExtractModuleDepsKotlin(t *testing.T) {
	deps, err := extractKotlinFTLImports("test", "testdata/alphakotlin")
	assert.NoError(t, err)
	assert.Equal(t, []string{"builtin", "other"}, deps)
}
