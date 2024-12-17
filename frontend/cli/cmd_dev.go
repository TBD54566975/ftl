package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/types/optional"
	"golang.org/x/sync/errgroup"

	"github.com/block/ftl/backend/admin"
	provisionerconnect "github.com/block/ftl/backend/protos/xyz/block/ftl/provisioner/v1beta1/provisionerpbconnect"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/timeline"
	"github.com/block/ftl/internal/bind"
	"github.com/block/ftl/internal/buildengine"
	"github.com/block/ftl/internal/configuration"
	"github.com/block/ftl/internal/configuration/manager"
	"github.com/block/ftl/internal/dev"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/lsp"
	"github.com/block/ftl/internal/projectconfig"
	"github.com/block/ftl/internal/rpc"
	"github.com/block/ftl/internal/schema/schemaeventsource"
	"github.com/block/ftl/internal/terminal"
)

type devCmd struct {
	Watch          time.Duration     `help:"Watch template directory at this frequency and regenerate on change." default:"500ms"`
	NoServe        bool              `help:"Do not start the FTL server." default:"false"`
	Lsp            bool              `help:"Run the language server." default:"false"`
	ServeCmd       serveCommonConfig `embed:""`
	InitDB         bool              `help:"Initialize the database and exit." default:"false"`
	languageServer *lsp.Server
	Build          buildCmd `embed:""`
}

func (d *devCmd) Run(
	ctx context.Context,
	k *kong.Kong,
	kctx *kong.Context,
	cm *manager.Manager[configuration.Configuration],
	sm *manager.Manager[configuration.Secrets],
	projConfig projectconfig.Config,
	bindContext terminal.KongContextBinder,
	schemaEventSourceFactory func() schemaeventsource.EventSource,
	controllerClient ftlv1connect.ControllerServiceClient,
	provisionerClient provisionerconnect.ProvisionerServiceClient,
	timelineClient *timeline.Client,
	adminClient admin.Client,
	verbClient ftlv1connect.VerbServiceClient,
) error {
	startTime := time.Now()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	if len(d.Build.Dirs) == 0 {
		d.Build.Dirs = projConfig.AbsModuleDirs()
	}
	if len(d.Build.Dirs) == 0 {
		return errors.New("no directories specified")
	}

	terminal.LaunchEmbeddedConsole(ctx, k, bindContext, schemaEventSourceFactory())
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
		err := dev.SetupPostgres(ctx, optional.Some(d.ServeCmd.DatabaseImage), d.ServeCmd.DBPort, true)
		if err != nil {
			return fmt.Errorf("failed to setup database: %w", err)
		}
		return nil
	}
	statusManager := terminal.FromContext(ctx)
	defer statusManager.Close()
	starting := statusManager.NewStatus("\u001B[92mStarting FTL Server 🚀\u001B[39m")

	bindAllocator, err := bind.NewBindAllocator(d.ServeCmd.Bind, 1)
	if err != nil {
		return fmt.Errorf("could not create bind allocator: %w", err)
	}

	// Default to allowing all origins and headers for console requests in local dev mode.
	d.ServeCmd.Ingress.AllowOrigins = []*url.URL{{Scheme: "*", Host: "*"}}
	d.ServeCmd.Ingress.AllowHeaders = []string{"*"}

	devModeEndpointUpdates := make(chan dev.LocalEndpoint, 1)
	// cmdServe will notify this channel when startup commands are complete and the controller is ready
	controllerReady := make(chan bool, 1)
	if !d.NoServe {
		if d.ServeCmd.Stop {
			err := d.ServeCmd.run(ctx, projConfig, cm, sm, optional.Some(controllerReady), true, bindAllocator, controllerClient, provisionerClient, timelineClient, adminClient, schemaEventSourceFactory, verbClient, true, devModeEndpointUpdates)
			if err != nil {
				return fmt.Errorf("failed to stop server: %w", err)
			}
			d.ServeCmd.Stop = false
		}

		g.Go(func() error {
			err := d.ServeCmd.run(ctx, projConfig, cm, sm, optional.Some(controllerReady), true, bindAllocator, controllerClient, provisionerClient, timelineClient, adminClient, schemaEventSourceFactory, verbClient, true, devModeEndpointUpdates)
			cancel()
			return err
		})
	}

	g.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case <-controllerReady:
		}
		starting.Close()

		opts := []buildengine.Option{buildengine.Parallelism(d.Build.Parallelism), buildengine.BuildEnv(d.Build.BuildEnv), buildengine.WithDevMode(devModeEndpointUpdates), buildengine.WithStartTime(startTime)}
		if d.Lsp {
			d.languageServer = lsp.NewServer(ctx)
			ctx = log.ContextWithLogger(ctx, log.FromContext(ctx).AddSink(lsp.NewLogSink(d.languageServer)))
			g.Go(func() error {
				return d.languageServer.Run()
			})
		}

		engine, err := buildengine.New(ctx, client, schemaEventSourceFactory(), projConfig, d.Build.Dirs, opts...)
		if err != nil {
			return err
		}
		if d.languageServer != nil {
			d.languageServer.Subscribe(ctx, engine.EngineUpdates)
		}
		return engine.Dev(ctx, d.Watch)
	})

	err = g.Wait()
	if err != nil {
		return fmt.Errorf("error during dev: %w", err)
	}
	return nil
}
