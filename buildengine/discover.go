package buildengine

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/TBD54566975/ftl/common/moduleconfig"
)

// DiscoverModules recursively loads all modules under the given directories.
//
// If no directories are provided, the current working directory is used.
func DiscoverModules(ctx context.Context, dirs ...string) ([]moduleconfig.ModuleConfig, error) {
	if len(dirs) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		dirs = []string{cwd}
	}
	out := []moduleconfig.ModuleConfig{}
	for _, dir := range dirs {
		err := WalkDir(dir, func(path string, d fs.DirEntry) error {
			if filepath.Base(path) != "ftl.toml" {
				return nil
			}
			moduleDir := filepath.Dir(path)
			config, err := moduleconfig.LoadModuleConfig(moduleDir)
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
