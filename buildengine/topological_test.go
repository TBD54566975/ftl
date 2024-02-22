package buildengine

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/common/moduleconfig"
)

func TestBuildOrder(t *testing.T) {
	modules := []Module{
		{
			ModuleConfig: moduleconfig.ModuleConfig{Module: "alpha"},
			Dependencies: []string{"beta", "gamma"},
		},
		{
			ModuleConfig: moduleconfig.ModuleConfig{Module: "beta"},
			Dependencies: []string{"kappa"},
		},
		{
			ModuleConfig: moduleconfig.ModuleConfig{Module: "gamma"},
			Dependencies: []string{"kappa"},
		},
		{
			ModuleConfig: moduleconfig.ModuleConfig{Module: "kappa"},
		},
		{
			ModuleConfig: moduleconfig.ModuleConfig{Module: "delta"},
		},
	}

	graph, err := BuildOrder(modules)
	assert.NoError(t, err)

	expected := [][]Module{
		{
			{ModuleConfig: moduleconfig.ModuleConfig{Module: "delta"}},
			{ModuleConfig: moduleconfig.ModuleConfig{Module: "kappa"}},
		},
		{
			{ModuleConfig: moduleconfig.ModuleConfig{Module: "beta"}, Dependencies: []string{"kappa"}},
			{ModuleConfig: moduleconfig.ModuleConfig{Module: "gamma"}, Dependencies: []string{"kappa"}},
		},
		{
			{ModuleConfig: moduleconfig.ModuleConfig{Module: "alpha"}, Dependencies: []string{"beta", "gamma"}},
		},
	}
	assert.Equal(t, expected, graph)
}
