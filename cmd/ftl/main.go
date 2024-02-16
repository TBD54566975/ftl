package main

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	"connectrpc.com/connect"
	"github.com/alecthomas/kong"
	kongtoml "github.com/alecthomas/kong-toml"

	"github.com/TBD54566975/ftl"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	_ "github.com/TBD54566975/ftl/internal/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type CLI struct {
	Version   kong.VersionFlag `help:"Show version."`
	Config    kong.ConfigFlag  `help:"Load configuration from TOML file." placeholder:"FILE"`
	LogConfig log.Config       `embed:"" prefix:"log-" group:"Logging:"`
	Endpoint  *url.URL         `default:"http://127.0.0.1:8892" help:"FTL endpoint to bind/connect to." env:"FTL_ENDPOINT"`

	Authenticators map[string]string `help:"Authenticators to use for FTL endpoints." mapsep:"," env:"FTL_AUTHENTICATORS" placeholder:"HOST=EXE,…"`

	Status   statusCmd   `cmd:"" help:"Show FTL status."`
	Init     initCmd     `cmd:"" help:"Initialize a new FTL module."`
	Dev      devCmd      `cmd:"" help:"Watch a directory for FTL modules and hot reload them."`
	PS       psCmd       `cmd:"" help:"List deployments."`
	Serve    serveCmd    `cmd:"" help:"Start the FTL server."`
	Call     callCmd     `cmd:"" help:"Call an FTL function."`
	Update   updateCmd   `cmd:"" help:"Update a deployment."`
	Kill     killCmd     `cmd:"" help:"Kill a deployment."`
	Schema   schemaCmd   `cmd:"" help:"FTL schema commands."`
	Build    buildCmd    `cmd:"" help:"Build an FTL module."`
	Deploy   deployCmd   `cmd:"" help:"Create a new deployment."`
	Download downloadCmd `cmd:"" help:"Download a deployment."`
}

var cli CLI

func main() {
	kctx := kong.Parse(&cli,
		kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`),
		kong.Configuration(kongtoml.Loader, ".ftl.toml", "~/.ftl.toml"),
		kong.AutoGroup(func(parent kong.Visitable, flag *kong.Flag) *kong.Group {
			node, ok := parent.(*kong.Command)
			if !ok {
				return nil
			}
			return &kong.Group{Key: node.Name, Title: "Command flags:"}
		}),
		kong.Vars{
			"version": ftl.Version,
			"os":      runtime.GOOS,
			"arch":    runtime.GOARCH,
		},
	)

	rpc.InitialiseClients(cli.Authenticators)

	// Set the log level for child processes.
	os.Setenv("LOG_LEVEL", cli.LogConfig.Level.String())

	ctx, cancel := context.WithCancel(context.Background())

	logger := log.Configure(os.Stderr, cli.LogConfig)
	ctx = log.ContextWithLogger(ctx, logger)

	// Handle signals.
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigch
		logger.Debugf("FTL terminating with signal %s", sig)
		cancel()
		_ = syscall.Kill(-syscall.Getpid(), sig.(syscall.Signal)) //nolint:forcetypeassert
		os.Exit(0)
	}()

	kctx.Bind(cli.Endpoint)
	kctx.BindTo(ctx, (*context.Context)(nil))
	err := kctx.BindToProvider(makeDialer(ftlv1connect.NewVerbServiceClient))
	kctx.FatalIfErrorf(err)
	err = kctx.BindToProvider(makeDialer(ftlv1connect.NewControllerServiceClient))
	kctx.FatalIfErrorf(err)

	err = kctx.Run(ctx)
	kctx.FatalIfErrorf(err)
}

func makeDialer[Client rpc.Pingable](newClient func(connect.HTTPClient, string, ...connect.ClientOption) Client) func() (Client, error) {
	return func() (Client, error) {
		return rpc.Dial(newClient, cli.Endpoint.String(), log.Error), nil
	}
}

// gitRoot returns the root of the git repository containing dir, or empty string if dir is not in a git repository.
//
// If dir is empty, the current working directory is used.
func gitRoot(ctx context.Context, dir string) string {
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return ""
		}
	}
	gitRootBytes, err := exec.Capture(ctx, dir, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(gitRootBytes))
}
