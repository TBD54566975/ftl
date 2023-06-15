package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/go-runtime/compile"
	"github.com/TBD54566975/ftl/go-runtime/compile/generate"
	"github.com/TBD54566975/ftl/schema"
)

type goCmd struct {
	Schema   goSchemaCmd   `cmd:"" help:"Extract the FTL schema from a Go module."`
	Generate goGenerateCmd `cmd:"" help:"Generate Go stubs for a module."`
	Build    goBuildCmd    `cmd:"" help:"Compile a Go module into a deployable executable."`
}

type goSchemaCmd struct {
	Dir []string `arg:"" help:"Path to root directory of module." type:"existingdir"`
}

func (g *goSchemaCmd) Run() error {
	s := &schema.Schema{}
	for _, dir := range g.Dir {
		module, err := compile.ExtractModuleSchema(dir)
		if err != nil {
			return errors.WithStack(err)
		}
		s.Modules = append(s.Modules, module)
	}
	if err := schema.Validate(s); err != nil {
		return errors.WithStack(err)
	}
	fmt.Println(s)
	return nil
}

type goGenerateCmd struct {
	Schema *os.File `arg:"" required:"" help:"Path to FTL schema file." default:"-"`
}

func (g *goGenerateCmd) Run() error {
	s, err := schema.Parse("", g.Schema)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, module := range s.Modules {
		if err := generate.GenerateExternalModule(os.Stdout, module); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

type goBuildCmd struct {
	compile.Config
}

func (g *goBuildCmd) Run(ctx context.Context) error {
	deployment, err := compile.Compile(ctx, g.Config)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, artefact := range deployment.Artefacts {
		if named, ok := artefact.Content.(interface{ Name() string }); ok {
			fmt.Println(named.Name())
			continue
		}
	}
	return nil
}
