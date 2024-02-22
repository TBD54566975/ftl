package buildengine

import "sort"

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
		group := []ModuleConfig{}
		for _, module := range modulesByName {
			// Check that all dependencies have been built.
			allBuilt := true
			for _, dep := range module.Dependencies {
				if !built[dep] {
					allBuilt = false
					break
				}
			}
			if allBuilt {
				group = append(group, module)
				built[module.Module] = true
				delete(modulesByName, module.Module)
			}
		}
		sort.Slice(group, func(i, j int) bool {
			return group[i].Module < group[j].Module
		})
		groups = append(groups, group)
	}
	return groups, nil
}
