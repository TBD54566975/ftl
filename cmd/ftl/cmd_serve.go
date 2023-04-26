package main

import (
	"context"

	"github.com/alecthomas/errors"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/agent"
	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/common/socket"
)

type serveCmd struct {
	Dir []string `arg:"" help:"Paths to FTL modules."`
}

func (r *serveCmd) Run(ctx context.Context, s socket.Socket) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger := log.FromContext(ctx).Sub("agent", log.Default)
	logger.Warnf("Starting FTL local agent on http://%s", s.Addr)
	ctx = log.ContextWithLogger(ctx, logger)

	agent, err := agent.New(ctx, s)
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
