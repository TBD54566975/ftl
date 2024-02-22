package main

import (
	"archive/zip"
	"bufio"
	"context"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/TBD54566975/scaffolder"
	"github.com/beevik/etree"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/buildengine"
	goruntime "github.com/TBD54566975/ftl/go-runtime"
	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	kotlinruntime "github.com/TBD54566975/ftl/kotlin-runtime"
)

type initCmd struct {
	Hermit bool          `help:"Include Hermit language-specific toolchain binaries in the module." negatable:""`
	Go     initGoCmd     `cmd:"" help:"Initialize a new FTL Go module."`
	Kotlin initKotlinCmd `cmd:"" help:"Initialize a new FTL Kotlin module."`
}

type initGoCmd struct {
	Replace map[string]string `short:"r" help:"Replace a module import path with a local path in the initialised FTL module." placeholder:"OLD=NEW,..."`
	Dir     string            `arg:"" help:"Directory to initialize the module in."`
	Name    string            `arg:"" help:"Name of the FTL module to create underneath the base directory."`
}

func (i initGoCmd) Run(ctx context.Context, parent *initCmd) error {
	if i.Name == "" {
		i.Name = filepath.Base(i.Dir)
	}
	if !schema.ValidateName(i.Name) {
		return fmt.Errorf("module name %q is invalid", i.Name)
	}
	logger := log.FromContext(ctx)
	logger.Debugf("Initializing FTL Go module %s in %s", i.Name, i.Dir)
	if err := scaffold(parent.Hermit, goruntime.Files(), i.Dir, i, scaffolder.Exclude("^go.mod$")); err != nil {
		return err
	}
	if err := updateGitIgnore(i.Dir); err != nil {
		return err
	}
	logger.Debugf("Running go mod tidy")

	return exec.Command(ctx, log.Debug, filepath.Join(i.Dir, i.Name), "go", "mod", "tidy").RunBuffered(ctx)
}

type initKotlinCmd struct {
	GroupID    string `short:"g" help:"Base Maven group ID (defaults to \"ftl\")." default:"ftl"`
	ArtifactID string `short:"a" help:"Base Maven artifact ID (defaults to \"ftl\")." default:"ftl"`
	Dir        string `arg:"" help:"Directory to initialize the module in."`
	Name       string `arg:"" help:"Name of the FTL module to create underneath the base directory."`
}

func (i initKotlinCmd) Run(ctx context.Context, parent *initCmd) error {
	if i.Name == "" {
		i.Name = filepath.Base(i.Dir)
	}

	if !schema.ValidateName(i.Name) {
		return fmt.Errorf("module name %q is invalid", i.Name)
	}

	if _, err := os.Stat(filepath.Join(i.Dir, "ftl-module-"+i.Name)); err == nil {
		return fmt.Errorf("module directory %s already exists", filepath.Join(i.Dir, i.Name))
	}

	options := []scaffolder.Option{}

	// Update root POM if it already exists.
	pomFile := filepath.Join(i.Dir, "pom.xml")
	if _, err := os.Stat(pomFile); err == nil {
		options = append(options, scaffolder.Exclude("^pom.xml$"))
		if err := updatePom(pomFile, i.Name); err != nil {
			return err
		}
	}

	if err := scaffold(parent.Hermit, kotlinruntime.Files(), i.Dir, i, options...); err != nil {
		return err
	}

	return buildengine.SetPOMProperties(ctx, i.Dir)
}

func updatePom(pomFile, name string) error {
	tree := etree.NewDocument()
	err := tree.ReadFromFile(pomFile)
	if err != nil {
		return err
	}

	// Add new module entry to root of XML file
	root := tree.Root()
	modules := root.SelectElement("modules")
	if modules == nil {
		modules = root.CreateElement("modules")
	}
	modules.CreateText("    ")
	module := modules.CreateElement("module")
	module.SetText("ftl-module-" + name)
	modules.CreateText("\n    ")

	// Write updated XML file back to disk
	err = tree.WriteToFile(pomFile)
	if err != nil {
		return err
	}
	return nil
}

func unzipToTmpDir(reader *zip.Reader) (string, error) {
	tmpDir, err := os.MkdirTemp("", "ftl-init-*")
	if err != nil {
		return "", err
	}
	err = internal.UnzipDir(reader, tmpDir)
	if err != nil {
		return "", err
	}
	return tmpDir, nil
}

func scaffold(hermit bool, source *zip.Reader, destination string, ctx any, options ...scaffolder.Option) error {
	opts := []scaffolder.Option{scaffolder.Functions(scaffoldFuncs), scaffolder.Exclude("^go.mod$")}
	if !hermit {
		opts = append(opts, scaffolder.Exclude("^bin"))
	}
	opts = append(opts, options...)
	if err := internal.ScaffoldZip(source, destination, ctx, opts...); err != nil {
		return fmt.Errorf("%s: %w", "failed to scaffold", err)
	}
	return nil
}

var scaffoldFuncs = template.FuncMap{
	"snake":          strcase.ToLowerSnake,
	"screamingSnake": strcase.ToUpperSnake,
	"camel":          strcase.ToUpperCamel,
	"lowerCamel":     strcase.ToLowerCamel,
	"kebab":          strcase.ToLowerKebab,
	"screamingKebab": strcase.ToUpperKebab,
	"upper":          strings.ToUpper,
	"lower":          strings.ToLower,
	"title":          strings.Title,
	"typename":       schema.TypeName,
}

func updateGitIgnore(dir string) error {
	gitRoot := internal.GitRoot(dir)
	f, err := os.OpenFile(path.Join(gitRoot, ".gitignore"), os.O_RDWR|os.O_CREATE, 0644) //nolint:gosec
	if err != nil {
		return err
	}
	defer f.Close() //nolint:gosec

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "**/_ftl" {
			return nil
		}
	}

	if scanner.Err() != nil {
		return scanner.Err()
	}

	// append if not already present
	_, err = f.WriteString("**/_ftl\n")
	return err
}
