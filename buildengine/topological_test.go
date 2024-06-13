package buildengine

import (
	"sort"
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
	topo, unsorted := TopologicalSort(graph)
	expected := [][]string{
		{"delta", "kappa"},
		{"beta", "gamma"},
		{"alpha"},
	}
	assert.Equal(t, expected, topo)
	assert.Equal(t, nil, unsorted)
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
	topo, unsorted := TopologicalSort(graph)
	expected := [][]string{
		{"base"},
		{"kappa"},
		{"gamma"},
	}
	assert.Equal(t, expected, topo)
	sort.Strings(unsorted)
	assert.Equal(t, []string{"alpha", "beta", "delta"}, unsorted)
}
