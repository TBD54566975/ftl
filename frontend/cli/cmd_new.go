package main

import (
	"context"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"regexp"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/buildengine"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/projectconfig"
)

type newCmd struct {
	Language string `arg:"" help:"Language of the module to create."`
	Dir      string `arg:"" help:"Directory to initialize the module in."`
	Name     string `arg:"" help:"Name of the FTL module to create underneath the base directory."`
}

func (i newCmd) Run(ctx context.Context, config projectconfig.Config) error {
	name, path, err := validateModule(i.Dir, i.Name)
	if err != nil {
		return err
	}

	// Validate the module name with custom validation
	if !isValidModuleName(name) {
		return fmt.Errorf("module name %q must be a valid Go module name and not a reserved keyword", name)
	}

	logger := log.FromContext(ctx)
	logger.Debugf("Creating FTL %s module %q in %s", i.Language, name, path)

	moduleConfig := moduleconfig.ModuleConfig{
		Module:   name,
		Language: i.Language,
	}.Abs()
	plugin, err := buildengine.PluginFromConfig(ctx, moduleConfig, config.Root())
	plugin.CreateModule(ctx, moduleConfig)
	if err != nil {
		return err
	}

	_, ok := internal.GitRoot(i.Dir).Get()
	if !config.NoGit && ok {
		logger.Debugf("Adding files to git")
		if config.Hermit {
			if err := maybeGitAdd(ctx, i.Dir, "bin/*"); err != nil {
				return err
			}
		}
		if err := maybeGitAdd(ctx, i.Dir, filepath.Join(i.Name, "*")); err != nil {
			return err
		}
	}
	return nil
}

func validateModule(dir string, name string) (string, string, error) {
	if dir == "" {
		return "", "", fmt.Errorf("directory is required")
	}
	if name == "" {
		name = filepath.Base(dir)
	}
	if !schema.ValidateName(name) {
		return "", "", fmt.Errorf("module name %q is invalid", name)
	}
	path := filepath.Join(dir, name)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", "", fmt.Errorf("could not make %q an absolute path: %w", path, err)
	}
	if _, err := os.Stat(absPath); err == nil {
		return "", "", fmt.Errorf("module directory %s already exists", path)
	}
	return name, absPath, nil
}

func isValidModuleName(name string) bool {
	validNamePattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
	if !validNamePattern.MatchString(name) {
		return false
	}
	if token.Lookup(name).IsKeyword() {
		return false
	}
	return true
}
