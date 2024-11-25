package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/ingress"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	_ "github.com/TBD54566975/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/observability"
	"github.com/TBD54566975/ftl/internal/rpc"
)

var cli struct {
	Version             kong.VersionFlag     `help:"Show version."`
	ObservabilityConfig observability.Config `embed:"" prefix:"o11y-"`
	LogConfig           log.Config           `embed:"" prefix:"log-"`
	HTTPIngressConfig   ingress.Config       `embed:""`
	ConfigFlag          string               `name:"config" short:"C" help:"Path to FTL project cf file." env:"FTL_CONFIG" placeholder:"FILE"`
	ControllerEndpoint  *url.URL             `name:"ftl-endpoint" help:"Controller endpoint." env:"FTL_ENDPOINT" default:"http://127.0.0.1:8892"`
}

func main() {
	t, err := strconv.ParseInt(ftl.Timestamp, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid timestamp %q: %s", ftl.Timestamp, err))
	}
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - HTTP Ingress`),
		kong.UsageOnError(),
		kong.Vars{"version": ftl.Version, "timestamp": time.Unix(t, 0).Format(time.RFC3339)},
	)

	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err = observability.Init(ctx, false, "", "ftl-cron", ftl.Version, cli.ObservabilityConfig)
	kctx.FatalIfErrorf(err, "failed to initialize observability")

	verbClient := rpc.Dial(ftlv1connect.NewVerbServiceClient, cli.ControllerEndpoint.String(), log.Error)
	schemaClient := rpc.Dial(ftlv1connect.NewSchemaServiceClient, cli.ControllerEndpoint.String(), log.Error)

	err = ingress.Start(ctx, cli.HTTPIngressConfig, schemaClient, verbClient)
	kctx.FatalIfErrorf(err, "failed to start HTTP ingress")
}
