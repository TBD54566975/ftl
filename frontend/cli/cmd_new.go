package main

import (
	"context"
	"fmt"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/buildengine"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/slices"
)

type newCmd struct {
	Language string `arg:"" help:"Language of the module to create."`
	Dir      string `arg:"" help:"Directory to initialize the module in."`
	Name     string `arg:"" help:"Name of the FTL module to create underneath the base directory."`
}

// prepareNewCmd adds language specific flags to kong
// This allows the new command to have good support for language specific flags like:
// - help text (ftl new go --help)
// - default values
// - environment variable overrides
func prepareNewCmd(ctx context.Context, k *kong.Kong, args []string) error {
	if len(args) < 2 {
		return nil
	} else if args[0] != "new" {
		return nil
	}
	language := args[1]
	if len(language) == 0 {
		return nil
	}

	newCmdNode, ok := slices.Find(k.Model.Children, func(n *kong.Node) bool {
		return n.Name == "new"
	})
	if !ok {
		return fmt.Errorf("could not find new command")
	}

	plugin, err := buildengine.PluginFromConfig(ctx, moduleconfig.ModuleConfig{
		Language: language,
	}, "")
	if err != nil {
		return fmt.Errorf("could not create plugin for %v: %w", language, err)
	}

	flags, err := plugin.GetCreateModuleFlags(ctx)
	if err != nil {
		return fmt.Errorf("could not get CLI flags for %v plugin: %w", language, err)
	}

	registry := kong.NewRegistry().RegisterDefaults()
	for _, flag := range flags {
		var str string
		strPtr := &str
		flag.Target = reflect.ValueOf(strPtr).Elem()
		flag.Mapper = registry.ForValue(flag.Target)
		flag.Group = &kong.Group{
			Title: "Flags for " + strings.ToTitle(language[0:1]) + language[1:] + " modules",
			Key:   "languageSpecificFlags",
		}
	}
	newCmdNode.Flags = append(newCmdNode.Flags, flags...)
	return nil
}

func (i newCmd) Run(ctx context.Context, ktctx *kong.Context, config projectconfig.Config) error {
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
		Dir:      path,
	}

	flags := map[string]string{}
	for _, f := range ktctx.Selected().Flags {
		flagValue, ok := f.Target.Interface().(string)
		if !ok {
			return fmt.Errorf("expected %v value to be a string but it was %T", f.Name, f.Target.Interface())
		}
		flags[f.Name] = flagValue
	}

	plugin, err := buildengine.PluginFromConfig(ctx, moduleConfig, config.Root())
	if err != nil {
		return err
	}
	err = plugin.CreateModule(ctx, config, moduleConfig, flags)
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
