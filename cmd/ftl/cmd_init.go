package main

import (
	"archive/zip"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/TBD54566975/scaffolder"
	"github.com/beevik/etree"
	"github.com/iancoleman/strcase"

	goruntime "github.com/TBD54566975/ftl/go-runtime"
	"github.com/TBD54566975/ftl/internal"
	kotlinruntime "github.com/TBD54566975/ftl/kotlin-runtime"
)

type initCmd struct {
	Hermit bool          `default:"true" help:"Include Hermit language-specific toolchain binaries in the module." negatable:""`
	Go     initGoCmd     `cmd:"" help:"Initialize a new FTL Go module."`
	Kotlin initKotlinCmd `cmd:"" help:"Initialize a new FTL Kotlin module."`
}

type initGoCmd struct {
	Dir  string `arg:"" help:"Directory to initialize the module in."`
	Name string `arg:"" help:"Name of the FTL module to create underneath the base directory."`
}

func (i initGoCmd) Run(parent *initCmd) error {
	if i.Name == "" {
		i.Name = filepath.Base(i.Dir)
	}
	tmpDir, err := unzipToTmpDir(goruntime.Files)
	if err != nil {
		return fmt.Errorf("%s: %w", "failed to unzip kotlin runtime", err)
	}
	defer os.RemoveAll(tmpDir)

	return scaffold(parent.Hermit, tmpDir, i.Dir, i, scaffolder.Exclude("^go.mod$"))
}

type initKotlinCmd struct {
	GroupID    string `short:"g" help:"Base Maven group ID (defaults to \"ftl\")." default:"ftl"`
	ArtifactID string `short:"a" help:"Base Maven artifact ID (defaults to \"ftl\")." default:"ftl"`
	Dir        string `arg:"" help:"Directory to initialize the module in."`
	Name       string `arg:"" help:"Name of the FTL module to create underneath the base directory."`
}

func (i initKotlinCmd) Run(parent *initCmd) error {
	if i.Name == "" {
		i.Name = filepath.Base(i.Dir)
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

	tmpDir, err := unzipToTmpDir(kotlinruntime.Files)
	if err != nil {
		return fmt.Errorf("%s: %w", "failed to unzip kotlin runtime", err)
	}
	defer os.RemoveAll(tmpDir)

	return scaffold(parent.Hermit, tmpDir, i.Dir, i, options...)
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

func scaffold(hermit bool, source string, destination string, ctx any, options ...scaffolder.Option) error {
	opts := []scaffolder.Option{scaffolder.Functions(scaffoldFuncs), scaffolder.Exclude("^go.mod$")}
	if !hermit {
		opts = append(opts, scaffolder.Exclude("^bin"))
	}
	opts = append(opts, options...)
	if err := scaffolder.Scaffold(source, destination, ctx, opts...); err != nil {
		return fmt.Errorf("%s: %w", "failed to scaffold", err)
	}
	return nil
}

var scaffoldFuncs = template.FuncMap{
	"snake":          strcase.ToSnake,
	"screamingSnake": strcase.ToScreamingSnake,
	"camel":          strcase.ToCamel,
	"lowerCamel":     strcase.ToLowerCamel,
	"kebab":          strcase.ToKebab,
	"screamingKebab": strcase.ToScreamingKebab,
	"upper":          strings.ToUpper,
	"lower":          strings.ToLower,
	"title":          strings.Title,
	"typename": func(v any) string {
		return reflect.Indirect(reflect.ValueOf(v)).Type().Name()
	},
}
