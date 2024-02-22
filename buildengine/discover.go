package buildengine

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// DiscoverModules recursively loads all modules under the given directories.
//
// If no directories are provided, the current working directory is used.
func DiscoverModules(dirs ...string) ([]ModuleConfig, error) {
	if len(dirs) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		dirs = []string{cwd}
	}
	out := []ModuleConfig{}
	for _, dir := range dirs {
		err := WalkDir(dir, func(path string, d fs.DirEntry) error {
			if filepath.Base(path) != "ftl.toml" {
				return nil
			}
			moduleDir := filepath.Dir(path)
			config, err := LoadModuleConfig(moduleDir)
			if err != nil {
				return err
			}
			out = append(out, config)
			return ErrSkip
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Module < out[j].Module
	})
	return out, nil
}
