package buildengine

import (
	"sort"

	"golang.org/x/exp/maps"
)

// TopologicalSort returns a sequence of groups of modules in topological order
// that may be built in parallel.
func TopologicalSort(graph map[ProjectKey][]ProjectKey) [][]ProjectKey {
	modulesByKey := map[ProjectKey]bool{}
	for module := range graph {
		modulesByKey[module] = true
	}
	// Order of modules to build.
	// Each element is a list of modules that can be built in parallel.
	groups := [][]ProjectKey{}

	// Modules that have already been "built"
	built := map[ProjectKey]bool{"builtin": true}

	for len(modulesByKey) > 0 {
		// Current group of modules that can be built in parallel.
		group := map[ProjectKey]bool{}
	nextModule:
		for module := range modulesByKey {
			// Check that all dependencies have been built.
			for _, dep := range graph[module] {
				if !built[dep] {
					continue nextModule
				}
			}
			group[module] = true
			delete(modulesByKey, module)
		}
		orderedGroup := maps.Keys(group)

		sort.Slice(orderedGroup, func(i, j int) bool {
			return orderedGroup[i] < orderedGroup[j]
		})
		for _, module := range orderedGroup {
			built[module] = true
		}
		groups = append(groups, orderedGroup)
	}
	return groups
}
