package buildengine

import (
	"sort"

	"golang.org/x/exp/maps"
)

// BuildOrder returns groups of modules in topological order that can be built
// in parallel.
func BuildOrder(modules []Module) ([][]Module, error) {
	modulesByName := map[string]Module{}
	for _, module := range modules {
		modulesByName[module.Module] = module
	}
	// Order of modules to build.
	// Each element is a list of modules that can be built in parallel.
	groups := [][]Module{}

	// Modules that have already been "built"
	built := map[string]bool{}

	for len(modulesByName) > 0 {
		// Current group of modules that can be built in parallel.
		group := map[string]Module{}
	nextModule:
		for _, module := range modulesByName {
			// Check that all dependencies have been built.
			for _, dep := range module.Dependencies {
				if !built[dep] {
					continue nextModule
				}
			}
			group[module.Module] = module
			delete(modulesByName, module.Module)
		}
		orderedGroup := maps.Values(group)
		sort.Slice(orderedGroup, func(i, j int) bool {
			return orderedGroup[i].Module < orderedGroup[j].Module
		})
		for _, module := range orderedGroup {
			built[module.Module] = true
		}
		groups = append(groups, orderedGroup)
	}
	return groups, nil
}
