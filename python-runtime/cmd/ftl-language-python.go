package main

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/language/languagepbconnect"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	pythonplugin "github.com/TBD54566975/ftl/python-runtime/python-plugin"
)

type PythonPluginCLI struct {
	LogConfig log.Config `embed:"" prefix:"log-" group:"Logging:"`
	Bind      *url.URL   `required:"" help:"URL to bind to."`
}

func main() {
	var cli PythonPluginCLI
	kctx := kong.Parse(&cli,
		kong.Description(`Python language plugin for FTL`),
		kong.ShortUsageOnError(),
		kong.HelpOptions{Compact: true, WrapUpperBound: 80},
	)

	ctx, cancel := context.WithCancel(log.ContextWithNewDefaultLogger(context.Background()))
	logger := log.Configure(os.Stderr, cli.LogConfig)

	logger.Infof("ftl-language-python starting up")

	// Handle signals.
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigch
		logger.Debugf("ftl-language-python terminating with signal %s", sig)
		cancel()
		_ = syscall.Kill(-syscall.Getpid(), sig.(syscall.Signal)) //nolint:forcetypeassert,errcheck // best effort
		os.Exit(0)
	}()

	kctx.BindTo(ctx, (*context.Context)(nil))

	svc := pythonplugin.New()
	logger.Infof("ftl-language-python starting to serve on %v", cli.Bind)
	err := rpc.Serve(ctx,
		cli.Bind,
		rpc.GRPC(languagepbconnect.NewLanguageServiceHandler, svc),
		rpc.PProf(),
	)
	logger.Errorf(err, "ftl-language-python stopped serving")
	kctx.FatalIfErrorf(err)
}
