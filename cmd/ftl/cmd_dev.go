package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alecthomas/types/optional"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/common/projectconfig"
	"github.com/TBD54566975/ftl/internal/buildengine"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/lsp"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type devCmd struct {
	Parallelism    int           `short:"j" help:"Number of modules to build in parallel." default:"${numcpu}"`
	Dirs           []string      `arg:"" help:"Base directories containing modules." type:"existingdir" optional:""`
	Watch          time.Duration `help:"Watch template directory at this frequency and regenerate on change." default:"500ms"`
	NoServe        bool          `help:"Do not start the FTL server." default:"false"`
	Lsp            bool          `help:"Run the language server." default:"false"`
	ServeCmd       serveCmd      `embed:""`
	InitDB         bool          `help:"Initialize the database and exit." default:"false"`
	languageServer *lsp.Server
}

func (d *devCmd) Run(ctx context.Context, projConfig projectconfig.Config) error {
	if len(d.Dirs) == 0 {
		d.Dirs = projConfig.AbsModuleDirs()
	}
	if len(d.Dirs) == 0 {
		return errors.New("no directories specified")
	}

	client := rpc.ClientFromContext[ftlv1connect.ControllerServiceClient](ctx)

	g, ctx := errgroup.WithContext(ctx)

	if d.NoServe && d.ServeCmd.Stop {
		logger := log.FromContext(ctx)
		return KillBackgroundServe(logger)
	}

	if d.InitDB {
		dsn, err := d.ServeCmd.setupDB(ctx, d.ServeCmd.DatabaseImage)
		if err != nil {
			return fmt.Errorf("failed to setup database: %w", err)
		}
		fmt.Println(dsn)
		return nil
	}

	// cmdServe will notify this channel when startup commands are complete and the controller is ready
	controllerReady := make(chan bool, 1)
	if !d.NoServe {
		if d.ServeCmd.Stop {
			err := d.ServeCmd.Run(ctx, projConfig)
			if err != nil {
				return err
			}
			d.ServeCmd.Stop = false
		}
		if d.ServeCmd.isRunning(ctx, client) {
			return errors.New(ftlRunningErrorMsg)
		}

		g.Go(func() error { return d.ServeCmd.run(ctx, projConfig, optional.Some(controllerReady)) })
	}

	g.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case <-controllerReady:
		}

		opts := []buildengine.Option{buildengine.Parallelism(d.Parallelism)}
		if d.Lsp {
			d.languageServer = lsp.NewServer(ctx)
			opts = append(opts, buildengine.WithListener(d.languageServer))
			ctx = log.ContextWithLogger(ctx, log.FromContext(ctx).AddSink(lsp.NewLogSink(d.languageServer)))
			g.Go(func() error {
				return d.languageServer.Run()
			})
		}

		engine, err := buildengine.New(ctx, client, projConfig.Root(), d.Dirs, opts...)
		if err != nil {
			return err
		}
		return engine.Dev(ctx, d.Watch)
	})

	return g.Wait()
}
