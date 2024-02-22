package buildengine

import (
	"sort"

	"golang.org/x/exp/maps"
)

// BuildOrder returns groups of modules in topological order that can be built
// in parallel.
//
// [ExtractDependencies] must have been called on each module before calling this function.
func BuildOrder(modules []ModuleConfig) ([][]ModuleConfig, error) {
	modulesByName := map[string]ModuleConfig{}
	for _, module := range modules {
		modulesByName[module.Module] = module
	}
	// Order of modules to build.
	// Each element is a list of modules that can be built in parallel.
	groups := [][]ModuleConfig{}

	// Modules that have already been "built"
	built := map[string]bool{}

	for len(modulesByName) > 0 {
		// Current group of modules that can be built in parallel.
		group := map[string]ModuleConfig{}
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
