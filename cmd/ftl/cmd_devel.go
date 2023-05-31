package main

import (
	"context"

	"github.com/alecthomas/errors"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/agent"
	"github.com/TBD54566975/ftl/common/log"
)

type develCmd struct {
	Dir []string `arg:"" help:"Paths to FTL modules."`
}

func (r *develCmd) Run(ctx context.Context, cli *CLI) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger := log.FromContext(ctx).Sub("agent", log.Default)
	logger.Warnf("Starting console on %s", cli.Endpoint)
	ctx = log.ContextWithLogger(ctx, logger)

	agent, err := agent.New(ctx, cli.Endpoint)
	if err != nil {
		return errors.WithStack(err)
	}

	// Spawn modules in parallel.
	spawnwg := errgroup.Group{}
	for _, dir := range r.Dir {
		dir := dir
		spawnwg.Go(func() error { return agent.Manage(ctx, dir) })
	}
	err = spawnwg.Wait()
	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(agent.Serve(ctx))
}
