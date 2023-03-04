package main

import (
	context "context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/TBD54566975/ftl/common/log"
	drivego "github.com/TBD54566975/ftl/drive-go"
	ftlv1 "github.com/TBD54566975/ftl/internal/gen/xyz/block/ftl/v1"
)

var cli struct {
	LogConfig   log.Config     `embed:"" group:"Logging:"`
	DriveConfig drivego.Config `embed:"" group:"Drive:"`
	Socket      string         `required:"" env:"FTL_DRIVE_SOCKET" help:"Unix socket to listen on."`
}

func main() {
	kctx := kong.Parse(&cli, kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`))
	_ = os.RemoveAll(cli.Socket)

	ctx, cancel := context.WithCancel(context.Background())

	// Logging.
	logger := log.New(cli.LogConfig, os.Stderr).With("C", "FTL.drive-go")
	ctx = log.ContextWithLogger(ctx, logger)

	// Signal handling.
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigch
		logger.Info("FTL.drive-go terminating", "signal", sig)
		cancel()
		_ = syscall.Kill(-syscall.Getpid(), sig.(syscall.Signal)) //nolint:forcetypeassert
		os.Exit(0)
	}()

	// Listen on the socket.
	l, err := (&net.ListenConfig{}).Listen(ctx, "unix", cli.Socket)
	kctx.FatalIfErrorf(err)

	logger.Info("Starting FTL.drive-go server", "socket", cli.Socket)

	// Configure and start the gRPC server.
	gs := grpc.NewServer(
		grpc.ChainUnaryInterceptor(log.UnaryGRPCInterceptor(logger)),
		grpc.ChainStreamInterceptor(log.StreamGRPCInterceptor(logger)),
	)
	reflection.Register(gs)
	srv, err := drivego.New(ctx, cli.DriveConfig)
	kctx.FatalIfErrorf(err)
	ftlv1.RegisterDriveServiceServer(gs, srv)
	err = gs.Serve(l)
	kctx.FatalIfErrorf(err)
}
