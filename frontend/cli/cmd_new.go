package main

import (
	"archive/zip"
	"context"
	"fmt"
	"go/token"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/TBD54566975/scaffolder"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	goruntime "github.com/TBD54566975/ftl/go-runtime"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/jvm-runtime/java"
	"github.com/TBD54566975/ftl/jvm-runtime/kotlin"
)

type newCmd struct {
	Go     newGoCmd     `cmd:"" help:"Initialize a new FTL Go module."`
	Java   newJavaCmd   `cmd:"" help:"Initialize a new FTL Java module."`
	Kotlin newKotlinCmd `cmd:"" help:"Initialize a new FTL Kotlin module."`
}

type newGoCmd struct {
	Replace   map[string]string `short:"r" help:"Replace a module import path with a local path in the initialised FTL module." placeholder:"OLD=NEW,..." env:"FTL_INIT_GO_REPLACE"`
	Dir       string            `arg:"" help:"Directory to initialize the module in."`
	Name      string            `arg:"" help:"Name of the FTL module to create underneath the base directory."`
	GoVersion string
}

type newJavaCmd struct {
	Dir   string `arg:"" help:"Directory to initialize the module in."`
	Name  string `arg:"" help:"Name of the FTL module to create underneath the base directory."`
	Group string `help:"The Maven groupId of the project." default:"com.example"`
}
type newKotlinCmd struct {
	Dir   string `arg:"" help:"Directory to initialize the module in."`
	Name  string `arg:"" help:"Name of the FTL module to create underneath the base directory."`
	Group string `help:"The Maven groupId of the project." default:"com.example"`
}

func (i newGoCmd) Run(ctx context.Context, config projectconfig.Config) error {
	name, path, err := validateModule(i.Dir, i.Name)
	if err != nil {
		return err
	}

	// Validate the module name with custom validation
	if !isValidGoModuleName(name) {
		return fmt.Errorf("module name %q must be a valid Go module name and not a reserved keyword", name)
	}

	logger := log.FromContext(ctx)
	logger.Debugf("Creating FTL Go module %q in %s", name, path)

	i.GoVersion = runtime.Version()[2:]
	if err := scaffold(ctx, config.Hermit, goruntime.Files(), i.Dir, i, scaffolder.Exclude("^go.mod$")); err != nil {
		return err
	}

	logger.Debugf("Running go mod tidy")
	if err := exec.Command(ctx, log.Debug, path, "go", "mod", "tidy").RunBuffered(ctx); err != nil {
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

func (i newJavaCmd) Run(ctx context.Context, config projectconfig.Config) error {
	return RunJvmScaffolding(ctx, config, i.Dir, i.Name, i.Group, java.Files())
}

func (i newKotlinCmd) Run(ctx context.Context, config projectconfig.Config) error {
	return RunJvmScaffolding(ctx, config, i.Dir, i.Name, i.Group, kotlin.Files())
}

func RunJvmScaffolding(ctx context.Context, config projectconfig.Config, dir string, name string, group string, source *zip.Reader) error {
	name, path, err := validateModule(dir, name)
	if err != nil {
		return err
	}

	logger := log.FromContext(ctx)
	logger.Debugf("Creating FTL module %q in %s", name, path)

	packageDir := strings.ReplaceAll(group, ".", "/")

	javaContext := struct {
		Dir        string
		Name       string
		Group      string
		PackageDir string
	}{
		Dir:        dir,
		Name:       name,
		Group:      group,
		PackageDir: packageDir,
	}

	if err := scaffold(ctx, config.Hermit, source, dir, javaContext); err != nil {
		return err
	}

	_, ok := internal.GitRoot(dir).Get()
	if !config.NoGit && ok {
		logger.Debugf("Adding files to git")
		if config.Hermit {
			if err := maybeGitAdd(ctx, dir, "bin/*"); err != nil {
				return err
			}
		}
		if err := maybeGitAdd(ctx, dir, filepath.Join(name, "*")); err != nil {
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
	if _, err := os.Stat(path); err == nil {
		return "", "", fmt.Errorf("module directory %s already exists", path)
	}
	return name, path, nil
}

func isValidGoModuleName(name string) bool {
	validNamePattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
	if !validNamePattern.MatchString(name) {
		return false
	}
	if token.Lookup(name).IsKeyword() {
		return false
	}
	return true
}

func scaffold(ctx context.Context, includeBinDir bool, source *zip.Reader, destination string, sctx any, options ...scaffolder.Option) error {
	logger := log.FromContext(ctx)
	opts := []scaffolder.Option{scaffolder.Functions(scaffoldFuncs), scaffolder.Exclude("^go.mod$")}
	if !includeBinDir {
		logger.Debugf("Excluding bin directory")
		opts = append(opts, scaffolder.Exclude("^bin"))
	}
	opts = append(opts, options...)
	if err := internal.ScaffoldZip(source, destination, sctx, opts...); err != nil {
		return fmt.Errorf("failed to scaffold: %w", err)
	}
	return nil
}

var scaffoldFuncs = template.FuncMap{
	"snake":          strcase.ToLowerSnake,
	"screamingSnake": strcase.ToUpperSnake,
	"camel":          strcase.ToUpperCamel,
	"lowerCamel":     strcase.ToLowerCamel,
	"strippedCamel":  strcase.ToUpperStrippedCamel,
	"kebab":          strcase.ToLowerKebab,
	"screamingKebab": strcase.ToUpperKebab,
	"upper":          strings.ToUpper,
	"lower":          strings.ToLower,
	"title":          strings.Title,
	"typename":       schema.TypeName,
}