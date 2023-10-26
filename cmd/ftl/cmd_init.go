package main

import (
	"archive/zip"
	"os"
	"path/filepath"

	"github.com/alecthomas/errors"

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
	Dir  string `arg:"" default:"." help:"Directory to initialize the module in."`
	Name string `short:"n" help:"Name of the FTL module (defaults to name of directory)."`
}

func (i *initKotlinCmd) Run(parent *initCmd) error {
	if i.Name == "" {
		i.Name = filepath.Base(i.Dir)
	}
	return errors.WithStack(scaffold(kotlinruntime.Files, parent.Hermit, i.Dir, i))
}

func scaffold(reader *zip.Reader, hermit bool, dir string, ctx any) error {
	err := internal.UnzipDir(reader, dir)
	if err != nil {
		return errors.WithStack(err)
	}
	if err := os.Remove(filepath.Join(dir, "go.mod")); err != nil {
		return errors.WithStack(err)
	}
	if err := internal.Scaffold(dir, ctx); err != nil {
		return errors.WithStack(err)
	}
	if !hermit {
		if err := os.RemoveAll(filepath.Join(dir, "bin")); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
