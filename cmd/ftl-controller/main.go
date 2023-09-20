package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/alecthomas/kong"

	_ "github.com/TBD54566975/ftl/backend/common/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/controller"
)

var version = "dev"
var timestamp = "0"

var cli struct {
	Version          kong.VersionFlag  `help:"Show version."`
	LogConfig        log.Config        `embed:"" prefix:"log-"`
	ControllerConfig controller.Config `embed:""`
}

func main() {
	t, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("invalid timestamp %q: %s", timestamp, err))
	}
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`),
		kong.UsageOnError(),
		kong.Vars{"version": version, "timestamp": time.Unix(t, 0).Format(time.RFC3339)},
	)
	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, cli.LogConfig))
	err = controller.Start(ctx, cli.ControllerConfig)
	kctx.FatalIfErrorf(err)
}
