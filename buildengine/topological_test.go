package buildengine

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestBuildOrder(t *testing.T) {
	modules := []Module{
		{
			ModuleConfig: ModuleConfig{Module: "alpha"},
			Dependencies: []string{"beta", "gamma"},
		},
		{
			ModuleConfig: ModuleConfig{Module: "beta"},
			Dependencies: []string{"kappa"},
		},
		{
			ModuleConfig: ModuleConfig{Module: "gamma"},
			Dependencies: []string{"kappa"},
		},
		{
			ModuleConfig: ModuleConfig{Module: "kappa"},
		},
		{
			ModuleConfig: ModuleConfig{Module: "delta"},
		},
	}

	graph, err := BuildOrder(modules)
	assert.NoError(t, err)

	expected := [][]Module{
		{
			{ModuleConfig: ModuleConfig{Module: "delta"}},
			{ModuleConfig: ModuleConfig{Module: "kappa"}},
		},
		{
			{ModuleConfig: ModuleConfig{Module: "beta"}, Dependencies: []string{"kappa"}},
			{ModuleConfig: ModuleConfig{Module: "gamma"}, Dependencies: []string{"kappa"}},
		},
		{
			{ModuleConfig: ModuleConfig{Module: "alpha"}, Dependencies: []string{"beta", "gamma"}},
		},
	}
	assert.Equal(t, expected, graph)
}
