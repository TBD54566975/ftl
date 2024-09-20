package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/buildengine"
	"github.com/TBD54566975/ftl/internal/console"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/lsp"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type devCmd struct {
	Watch          time.Duration `help:"Watch template directory at this frequency and regenerate on change." default:"500ms"`
	NoServe        bool          `help:"Do not start the FTL server." default:"false"`
	Lsp            bool          `help:"Run the language server." default:"false"`
	ServeCmd       serveCmd      `embed:""`
	InitDB         bool          `help:"Initialize the database and exit." default:"false"`
	languageServer *lsp.Server
	Build          buildCmd `embed:""`
}

func (d *devCmd) Run(ctx context.Context, k *kong.Kong, projConfig projectconfig.Config) error {

	console.LaunchEmbeddedConsole(ctx, k, projConfig, bindContext)
	if len(d.Build.Dirs) == 0 {
		d.Build.Dirs = projConfig.AbsModuleDirs()
	}
	if len(d.Build.Dirs) == 0 {
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
	sm := console.FromContext(ctx)
	starting := sm.NewStatus("\u001B[92mStarting FTL Server 🚀\u001B[39m")
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

		g.Go(func() error {
			return d.ServeCmd.run(ctx, projConfig, optional.Some(controllerReady), true)
		})
	}

	g.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case <-controllerReady:
		}
		starting.Close()

		opts := []buildengine.Option{buildengine.Parallelism(d.Build.Parallelism), buildengine.BuildEnv(d.Build.BuildEnv), buildengine.WithDevMode(true)}
		if d.Lsp {
			d.languageServer = lsp.NewServer(ctx)
			opts = append(opts, buildengine.WithListener(d.languageServer))
			ctx = log.ContextWithLogger(ctx, log.FromContext(ctx).AddSink(lsp.NewLogSink(d.languageServer)))
			g.Go(func() error {
				return d.languageServer.Run()
			})
		}

		engine, err := buildengine.New(ctx, client, projConfig.Root(), d.Build.Dirs, opts...)
		if err != nil {
			return err
		}
		return engine.Dev(ctx, d.Watch)
	})

	return g.Wait()
}
