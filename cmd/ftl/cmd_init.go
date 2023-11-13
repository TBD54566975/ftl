package main

import (
	"archive/zip"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/iancoleman/strcase"

	goruntime "github.com/TBD54566975/ftl/go-runtime"
	"github.com/TBD54566975/ftl/internal"
	kotlinruntime "github.com/TBD54566975/ftl/kotlin-runtime"
	"github.com/TBD54566975/scaffolder"
)

type initCmd struct {
	Hermit bool          `default:"true" help:"Include Hermit language-specific toolchain binaries in the module." negatable:""`
	Go     initGoCmd     `cmd:"" help:"Initialize a new FTL Go module."`
	Kotlin initKotlinCmd `cmd:"" help:"Initialize a new FTL Kotlin module."`
}

type initGoCmd struct {
	Dir  string `arg:"" default:"." type:"dir" help:"Directory to initialize the module in."`
	Name string `short:"n" help:"Name of the FTL module (defaults to name of directory)."`
}

func (i initGoCmd) Run(parent *initCmd) error {
	if i.Name == "" {
		i.Name = filepath.Base(i.Dir)
	}
	return errors.WithStack(scaffold(goruntime.Files, parent.Hermit, i.Dir, i))
}

type initKotlinCmd struct {
	GroupID    string `short:"g" help:"Base Maven group ID (defaults to \"ftl\")." default:"ftl"`
	ArtifactID string `short:"a" help:"Base Maven artifact ID (defaults to \"ftl\")." default:"ftl"`
	Dir        string `arg:"" help:"Directory to initialize the module in."`
	Name       string `arg:"" help:"Name of the FTL module to create underneath the base directory."`
}

func (i *initKotlinCmd) Run(parent *initCmd) error {
	if i.Name == "" {
		i.Name = filepath.Base(i.Dir)
	}
	return errors.WithStack(scaffold(kotlinruntime.Files, parent.Hermit, i.Dir, i))
}

func scaffold(reader *zip.Reader, hermit bool, dir string, ctx any) error {
	tmpDir, err := os.MkdirTemp("", "ftl-init-*")
	if err != nil {
		return errors.WithStack(err)
	}
	defer os.RemoveAll(tmpDir)
	err = internal.UnzipDir(reader, tmpDir)
	if err != nil {
		return errors.WithStack(err)
	}
	opts := []scaffolder.Option{scaffolder.Functions(scaffoldFuncs), scaffolder.Exclude("go.mod")}
	if !hermit {
		opts = append(opts, scaffolder.Exclude("bin"))
	}
	if err := scaffolder.Scaffold(tmpDir, dir, ctx, opts...); err != nil {
		return errors.Wrap(err, "failed to scaffold")
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
