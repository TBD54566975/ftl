package main

import (
	"context"
	"os"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl/internal/log"
)

var cli struct {
	log.Config

	Plan  PlanCmd  `cmd:""`
	Apply ApplyCmd `cmd:""`
}

type PlanCmd struct{}

func (p *PlanCmd) Run() {

}

type ApplyCmd struct{}

func (p *ApplyCmd) Run() {

}

func main() {
	kctx := kong.Parse(&cli)
	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.Config))

	err := kctx.Run(ctx)

	kctx.FatalIfErrorf(err)
}
