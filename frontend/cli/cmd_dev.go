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
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1beta1/provisioner/provisionerconnect"
	"github.com/TBD54566975/ftl/internal/bind"
	"github.com/TBD54566975/ftl/internal/buildengine"
	"github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/configuration/manager"
	"github.com/TBD54566975/ftl/internal/dev"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/lsp"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/terminal"
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

func (d *devCmd) Run(
	ctx context.Context,
	k *kong.Kong,
	cm *manager.Manager[configuration.Configuration],
	sm *manager.Manager[configuration.Secrets],
	projConfig projectconfig.Config,
	bindContext terminal.KongContextBinder,
) error {
	startTime := time.Now()
	if len(d.Build.Dirs) == 0 {
		d.Build.Dirs = projConfig.AbsModuleDirs()
	}
	if len(d.Build.Dirs) == 0 {
		return errors.New("no directories specified")
	}

	controllerClient := rpc.ClientFromContext[ftlv1connect.ControllerServiceClient](ctx)
	terminal.LaunchEmbeddedConsole(ctx, k, bindContext, controllerClient)

	var client buildengine.DeployClient = controllerClient
	if d.ServeCmd.Provisioners > 0 {
		client = rpc.ClientFromContext[provisionerconnect.ProvisionerServiceClient](ctx)
	}

	g, ctx := errgroup.WithContext(ctx)

	if d.NoServe && d.ServeCmd.Stop {
		logger := log.FromContext(ctx)
		return KillBackgroundServe(logger)
	}

	if d.InitDB {
		dsn, err := dev.SetupDB(ctx, d.ServeCmd.DatabaseImage, d.ServeCmd.DBPort, d.ServeCmd.Recreate)
		if err != nil {
			return fmt.Errorf("failed to setup database: %w", err)
		}
		fmt.Println(dsn)
		return nil
	}
	statusManager := terminal.FromContext(ctx)
	starting := statusManager.NewStatus("\u001B[92mStarting FTL Server 🚀\u001B[39m")

	bindAllocator, err := bind.NewBindAllocator(d.ServeCmd.Bind, 1)
	if err != nil {
		return fmt.Errorf("could not create bind allocator: %w", err)
	}

	// cmdServe will notify this channel when startup commands are complete and the controller is ready
	controllerReady := make(chan bool, 1)
	if !d.NoServe {
		if d.ServeCmd.Stop {
			err := d.ServeCmd.Run(ctx, cm, sm, projConfig)
			if err != nil {
				return err
			}
			d.ServeCmd.Stop = false
		}

		g.Go(func() error {
			return d.ServeCmd.run(ctx, projConfig, cm, sm, optional.Some(controllerReady), true, bindAllocator)
		})
	}

	g.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case <-controllerReady:
		}
		starting.Close()

		opts := []buildengine.Option{buildengine.Parallelism(d.Build.Parallelism), buildengine.BuildEnv(d.Build.BuildEnv), buildengine.WithDevMode(true), buildengine.WithStartTime(startTime)}
		if d.Lsp {
			d.languageServer = lsp.NewServer(ctx)
			ctx = log.ContextWithLogger(ctx, log.FromContext(ctx).AddSink(lsp.NewLogSink(d.languageServer)))
			g.Go(func() error {
				return d.languageServer.Run()
			})
		}

		engine, err := buildengine.New(ctx, client, projConfig, d.Build.Dirs, opts...)
		if err != nil {
			return err
		}
		if d.languageServer != nil {
			d.languageServer.Subscribe(ctx, engine.EngineUpdates)
		}
		return engine.Dev(ctx, d.Watch)
	})

	return g.Wait()
}
