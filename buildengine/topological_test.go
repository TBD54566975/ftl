package buildengine

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestTopologicalSort(t *testing.T) {
	graph := map[ProjectKey][]ProjectKey{
		"alpha": {"beta", "gamma"},
		"beta":  {"kappa"},
		"gamma": {"kappa"},
		"kappa": {},
		"delta": {},
	}
	topo := TopologicalSort(graph)
	expected := [][]ProjectKey{
		{"delta", "kappa"},
		{"beta", "gamma"},
		{"alpha"},
	}
	assert.Equal(t, expected, topo)
}
