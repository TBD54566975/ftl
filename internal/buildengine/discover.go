package buildengine

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/TBD54566975/ftl/internal/log"
)

// DiscoverModules recursively loads all modules under the given directories
// (or if none provided, the current working directory is used).
func DiscoverModules(ctx context.Context, moduleDirs []string) ([]Module, error) {
	out := []Module{}
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
func discoverModules(dirs ...string) ([]Module, error) {
	if len(dirs) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		dirs = []string{cwd}
	}
	out := []Module{}
	for _, dir := range dirs {
		err := WalkDir(dir, func(path string, d fs.DirEntry) error {
			if filepath.Base(path) != "ftl.toml" {
				return nil
			}
			moduleDir := filepath.Dir(path)
			module, err := LoadModule(moduleDir)
			if err != nil {
				return err
			}
			out = append(out, module)
			return ErrSkip
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Config.Module < out[j].Config.Module
	})
	return out, nil
}
