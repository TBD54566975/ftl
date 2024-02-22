package buildengine

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestBuildOrder(t *testing.T) {
	modules := []ModuleConfig{
		{
			Module:       "alpha",
			Dependencies: []string{"beta", "gamma"},
		},
		{
			Module:       "beta",
			Dependencies: []string{"kappa"},
		},
		{
			Module:       "gamma",
			Dependencies: []string{"kappa"},
		},
		{
			Module: "kappa",
		},
		{
			Module: "delta",
		},
	}

	graph, err := BuildOrder(modules)
	assert.NoError(t, err)

	expected := [][]ModuleConfig{
		{
			{Module: "delta"},
			{Module: "kappa"},
		},
		{
			{Module: "beta", Dependencies: []string{"kappa"}},
			{Module: "gamma", Dependencies: []string{"kappa"}},
		},
		{
			{Module: "alpha", Dependencies: []string{"beta", "gamma"}},
		},
	}
	assert.Equal(t, expected, graph)
}
