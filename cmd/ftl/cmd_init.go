package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/backend/common/exec"
	"github.com/TBD54566975/ftl/backend/common/log"
	goruntime "github.com/TBD54566975/ftl/go-runtime"
	"github.com/TBD54566975/ftl/internal"
)

type initCmd struct {
	Hermit bool          `default:"true" help:"Include Hermit language-specific toolchain binaries in the module." negatable:""`
	Go     initGoCmd     `cmd:"" help:"Initialize a new FTL Go module."`
	Kotlin initKotlinCmd `cmd:"" help:"Initialize a new FTL Kotlin module."`
}

type initGoCmd struct {
	Dir      string `arg:"" default:"." type:"dir" help:"Directory to initialize the module in."`
	Name     string `help:"Name of the FTL module (defaults to name of directory)."`
	GoModule string `required:"" help:"Go module path."`
}

func (i initGoCmd) Run(ctx context.Context, parent *initCmd) error {
	if i.Name == "" {
		i.Name = filepath.Base(i.Dir)
	}
	if err := internal.Scaffold(goruntime.Files, i.Dir, i); err != nil {
		return errors.WithStack(err)
	}
	if !parent.Hermit {
		if err := os.RemoveAll(filepath.Join(i.Dir, "bin")); err != nil {
			return errors.WithStack(err)
		}
	}
	if err := exec.Command(ctx, log.Info, i.Dir, "go", "mod", "tidy").Run(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

type initKotlinCmd struct {
	Dir string `arg:"" default:"." help:"Directory to initialize the module in."`
}

func (i *initKotlinCmd) Run() error {
	panic("??")
}
