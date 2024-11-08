package main

import (
	"context"
	"fmt"
	"net/url"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver
	"github.com/jpillora/backoff"
	"golang.org/x/sync/errgroup"

	"github.com/TBD54566975/ftl/backend/controller"
	"github.com/TBD54566975/ftl/backend/controller/scaling/localscaling"
	"github.com/TBD54566975/ftl/backend/controller/sql/databasetesting"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/bind"
	"github.com/TBD54566975/ftl/internal/buildengine"
	"github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/configuration/manager"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema"
)

type boxRunCmd struct {
	Recreate          bool          `help:"Recreate the database."`
	DSN               string        `help:"DSN for the database." default:"${boxdsn}" env:"FTL_CONTROLLER_DSN"`
	IngressBind       *url.URL      `help:"Bind address for the ingress server." default:"http://0.0.0.0:8891" env:"FTL_INGRESS_BIND"`
	Bind              *url.URL      `help:"Bind address for the FTL controller." default:"http://0.0.0.0:8892" env:"FTL_BIND"`
	RunnerBase        *url.URL      `help:"Base bind address for FTL runners." default:"http://127.0.0.1:8893" env:"FTL_RUNNER_BIND"`
	Dir               string        `arg:"" help:"Directory to scan for precompiled modules." default:"."`
	ControllerTimeout time.Duration `help:"Timeout for Controller start." default:"30s"`
}

func (b *boxRunCmd) Run(
	ctx context.Context,
	projConfig projectconfig.Config,
	sm *manager.Manager[configuration.Secrets],
	cm *manager.Manager[configuration.Configuration],
) error {
	_, err := databasetesting.CreateForDevel(ctx, b.DSN, b.Recreate)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	config := controller.Config{
		Bind:        b.Bind,
		IngressBind: b.IngressBind,
		Key:         model.NewLocalControllerKey(0),
		DSN:         b.DSN,
	}
	config.SetDefaults()

	// Start the controller.
	bindAllocator, err := bind.NewBindAllocator(b.RunnerBase, 0)
	if err != nil {
		return fmt.Errorf("failed to create runner port allocator: %w", err)
	}
	runnerScaling, err := localscaling.NewLocalScaling(bindAllocator, []*url.URL{b.Bind}, "", false)
	if err != nil {
		return fmt.Errorf("failed to create runner autoscaler: %w", err)
	}

	// Bring up the DB connection and DAL.
	conn, err := config.OpenDBAndInstrument()
	if err != nil {
		return fmt.Errorf("failed to bring up DB connection: %w", err)
	}

	wg := errgroup.Group{}
	wg.Go(func() error {
		return controller.Start(ctx, config, runnerScaling, cm, sm, conn, false)
	})

	// Wait for the controller to come up.
	client := ftlv1connect.NewControllerServiceClient(rpc.GetHTTPClient(b.Bind.String()), b.Bind.String())
	if err := rpc.Wait(ctx, backoff.Backoff{}, b.ControllerTimeout, client); err != nil {
		return fmt.Errorf("controller failed to start: %w", err)
	}

	engine, err := buildengine.New(ctx, client, projConfig, []string{b.Dir})
	if err != nil {
		return fmt.Errorf("failed to create build engine: %w", err)
	}

	logger := log.FromContext(ctx)

	// Manually import the schema for each module to get the dependency graph.
	err = engine.Each(func(m buildengine.Module) error {
		logger.Debugf("Loading schema for module %q", m.Config.Module)
		mod, err := schema.ModuleFromProtoFile(projConfig.SchemaPath(m.Config.Module))
		if err != nil {
			return fmt.Errorf("failed to read schema for module %q: %w", m.Config.Module, err)
		}
		engine.Import(ctx, mod)
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to load schemas: %w", err)
	}

	if err := engine.Deploy(ctx, 1, true); err != nil {
		return fmt.Errorf("failed to deploy: %w", err)
	}
	return wg.Wait()
}
