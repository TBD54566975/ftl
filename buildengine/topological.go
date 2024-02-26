package buildengine

import (
	"sort"

	"golang.org/x/exp/maps"
)

// TopologicalSort returns a sequence of groups of modules in topological order
// that may be built in parallel.
func TopologicalSort(graph map[string][]string) [][]string {
	modulesByName := map[string]bool{}
	for module := range graph {
		modulesByName[module] = true
	}
	// Order of modules to build.
	// Each element is a list of modules that can be built in parallel.
	groups := [][]string{}

	// Modules that have already been "built"
	built := map[string]bool{"builtin": true}

	for len(modulesByName) > 0 {
		// Current group of modules that can be built in parallel.
		group := map[string]bool{}
	nextModule:
		for module := range modulesByName {
			// Check that all dependencies have been built.
			for _, dep := range graph[module] {
				if !built[dep] {
					continue nextModule
				}
			}
			group[module] = true
			delete(modulesByName, module)
		}
		orderedGroup := maps.Keys(group)
		sort.Strings(orderedGroup)
		for _, module := range orderedGroup {
			built[module] = true
		}
		groups = append(groups, orderedGroup)
	}
	return groups
}
