package buildengine

import (
	"fmt"
	"sort"

	"golang.org/x/exp/maps"
)

// TopologicalSort attempts to order the modules supplied in the graph based on
// their topologically sorted order. A cycle in the module dependency graph
// will cause this sort to be incomplete. The sorted modules are returned as a
// sequence of `groups` of modules that may be built in parallel. The `unsorted`
// modules impacted by a dependency cycle are listed in random order.
func TopologicalSort(graph map[string][]string) (groups [][]string, unsorted []string) {
	modulesByName := map[string]bool{}
	for module := range graph {
		modulesByName[module] = true
	}

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
		// A module dependency cycle prevents further sorting
		if len(group) == 0 {
			// The remaining modules are either a member of the cyclical
			// dependency chain or depend (directly or transitively) on
			// a member of the cyclical dependency chain
			unsorted = maps.Keys(modulesByName)
			break
		}
		orderedGroup := maps.Keys(group)
		sort.Strings(orderedGroup)
		for _, module := range orderedGroup {
			built[module] = true
		}
		groups = append(groups, orderedGroup)
	}
	return groups, unsorted
}

func NewDependencyCycleError(unsorted []string) error {
	if len(unsorted) > 0 {
		return fmt.Errorf("detected a module dependency cycle that impacts these modules: %q", unsorted)
	}
	return nil
}
