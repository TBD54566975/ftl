package main

import (
	"context"
	"errors"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/buildengine"
	"github.com/TBD54566975/ftl/common/projectconfig"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type devCmd struct {
	Parallelism int           `short:"j" help:"Number of modules to build in parallel." default:"${numcpu}"`
	Dirs        []string      `arg:"" help:"Base directories containing modules." type:"existingdir" optional:""`
	External    []string      `help:"Directories for libraries that require FTL module stubs." type:"existingdir" optional:""`
	Watch       time.Duration `help:"Watch template directory at this frequency and regenerate on change." default:"500ms"`
	NoServe     bool          `help:"Do not start the FTL server." default:"false"`
	ServeCmd    serveCmd      `embed:""`
}

func (d *devCmd) Run(ctx context.Context, projConfig projectconfig.Config) error {
	if len(d.Dirs) == 0 && len(d.External) == 0 {
		d.Dirs = projConfig.ModuleDirs
		d.External = projConfig.ExternalDirs
	}
	if len(d.Dirs) == 0 && len(d.External) == 0 {
		return errors.New("no directories specified")
	}

	client := rpc.ClientFromContext[ftlv1connect.ControllerServiceClient](ctx)

	g, ctx := errgroup.WithContext(ctx)

	if !d.NoServe {
		if d.ServeCmd.Stop {
			err := d.ServeCmd.Run(ctx)
			if err != nil {
				return err
			}
			d.ServeCmd.Stop = false
		}
		if d.ServeCmd.isRunning(ctx, client) {
			return errors.New(ftlRunningErrorMsg)
		}

		g.Go(func() error {
			return d.ServeCmd.Run(ctx)
		})
	}

	err := d.ServeCmd.pollControllerOnine(ctx, client)
	if err != nil {
		return err
	}

	g.Go(func() error {
		engine, err := buildengine.New(ctx, client, d.Dirs, d.External, buildengine.Parallelism(d.Parallelism))
		if err != nil {
			return err
		}
		return engine.Dev(ctx, d.Watch)
	})

	return g.Wait()
}
