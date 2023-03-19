package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/schema"
	sdkgo "github.com/TBD54566975/ftl/sdk-go"
)

type goCmd struct {
	Schema   goSchemaCmd   `cmd:"" help:"Extract the FTL schema from a Go module."`
	Generate goGenerateCmd `cmd:"" help:"Generate Go stubs for a module."`
}

type goSchemaCmd struct {
	Dir []string `arg:"" help:"Path to root directory of module." type:"existingdir"`
}

func (g *goSchemaCmd) Run() error {
	s := schema.Schema{}
	for _, dir := range g.Dir {
		module, err := sdkgo.ExtractModule(dir)
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
		if err := sdkgo.Generate(module, os.Stdout); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}
