package watch

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/moduleconfig"
)

// DiscoverModules recursively loads all modules under the given directories
// (or if none provided, the current working directory is used).
func DiscoverModules(ctx context.Context, moduleDirs []string) ([]moduleconfig.UnvalidatedModuleConfig, error) {
	out := []moduleconfig.UnvalidatedModuleConfig{}
	logger := log.FromContext(ctx)

	modules, err := discoverModules(moduleDirs...)
	if err != nil {
		logger.Tracef("error discovering modules: %v", err)
		return nil, err
	}

	out = append(out, modules...)
	return out, nil
}

// discoverModules recursively loads all modules under the given directories.
//
// If no directories are provided, the current working directory is used.
func discoverModules(dirs ...string) ([]moduleconfig.UnvalidatedModuleConfig, error) {
	if len(dirs) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		dirs = []string{cwd}
	}
	out := []moduleconfig.UnvalidatedModuleConfig{}
	for _, dir := range dirs {
		err := WalkDir(dir, true, func(path string, d fs.DirEntry) error {
			if filepath.Base(path) != "ftl.toml" {
				return nil
			}
			moduleDir := filepath.Dir(path)
			config, err := moduleconfig.LoadConfig(moduleDir)
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
