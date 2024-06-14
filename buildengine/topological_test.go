package buildengine

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestTopologicalSort(t *testing.T) {
	graph := map[string][]string{
		"alpha": {"beta", "gamma"},
		"beta":  {"kappa"},
		"gamma": {"kappa"},
		"kappa": {},
		"delta": {},
	}
	topo, err := TopologicalSort(graph)
	expected := [][]string{
		{"delta", "kappa"},
		{"beta", "gamma"},
		{"alpha"},
	}
	assert.Equal(t, expected, topo)
	assert.NoError(t, err)
}

func TestTopologicalSortCycleDetection(t *testing.T) {
	graph := map[string][]string{
		"alpha": {"beta", "base"},
		"beta":  {"alpha", "base"},
		"delta": {"alpha", "base"},
		"kappa": {"base"},
		"gamma": {"kappa", "base"},
		"base":  {},
	}
	topo, err := TopologicalSort(graph)
	expected := [][]string{
		{"base"},
		{"kappa"},
		{"gamma"},
	}
	assert.Equal(t, expected, topo)
	assert.Error(t, err)
	assert.Equal(t, "detected a module dependency cycle that impacts these modules: [\"alpha\" \"beta\" \"delta\"]", err.Error())
}
