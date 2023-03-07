package plugin

import (
	"context"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/kong"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"github.com/TBD54566975/ftl/common/exec"
	ftlv1 "github.com/TBD54566975/ftl/common/gen/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/common/log"
)

// PingableClient is a gRPC client that can be pinged.
type PingableClient interface {
	Ping(ctx context.Context, in *ftlv1.PingRequest, opts ...grpc.CallOption) (*ftlv1.PingResponse, error)
}

type pluginOptions struct {
	envars []string
}

// Option used when creating a plugin.
type Option func(*pluginOptions)

// WithEnvars sets the environment variables to pass to the plugin.
func WithEnvars(envars ...string) Option {
	return func(o *pluginOptions) {
		o.envars = envars
	}
}

type Plugin[Client PingableClient] struct {
	Cmd    *exec.Cmd
	Client Client
}

// Spawn a new sub-process plugin.
//
// Plugins are gRPC servers that listen on a unix socket passed in an envar.
//
// The subprocess should call [Start] to start the gRPC server.
//
// "cmdCtx" will be cancelled when the plugin stops.
func Spawn[Client PingableClient](
	ctx context.Context,
	dir, exe string,
	makeClient func(grpc.ClientConnInterface) Client,
	options ...Option,
) (
	cmdCtx context.Context,
	plugin *Plugin[Client],
	err error,
) {
	name := "FTL." + filepath.Base(exe)
	logger := log.FromContext(ctx).With("C", name)

	opts := pluginOptions{}
	for _, opt := range options {
		opt(&opts)
	}
	workingDir := filepath.Join(dir, ".ftl")
	socket := filepath.Join(workingDir, filepath.Base(exe)+".sock")
	logger.Info("Spawning plugin", "dir", dir, "exe", exe, "socket", socket)
	cmd := exec.Command(ctx, dir, exe)
	cmd.Env = append(cmd.Env, "FTL_PLUGIN_SOCKET="+socket)
	cmd.Env = append(cmd.Env, "FTL_WORKING_DIR="+workingDir)
	cmd.Env = append(cmd.Env, opts.envars...)

	if err = cmd.Start(); err != nil {
		return nil, nil, errors.WithStack(err)
	}
	defer func() {
		if err != nil {
			_ = cmd.Kill(syscall.SIGKILL)
		}
	}()

	logger.Info("Dialing plugin")

	var cancelWithCause context.CancelCauseFunc
	// Cancel the context if the command exits - this will terminate the Dial immediately.
	cmdCtx, cancelWithCause = context.WithCancelCause(ctx)
	go func() { cancelWithCause(cmd.Wait()) }()

	conn, err := grpc.DialContext(ctx, "unix://"+socket,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}

	// Close the gRPC connection when the context is cancelled.
	go func() {
		<-cmdCtx.Done()
		_ = conn.Close()
	}()

	client := makeClient(conn)

	for i := 0; i < 30*10; i++ {
		_, err = client.Ping(ctx, &ftlv1.PingRequest{})
		if err == nil {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	if err != nil {
		return nil, nil, errors.Wrap(err, "plugin did not respond to ping")
	}

	logger.Info(name + " online")
	plugin = &Plugin[Client]{Cmd: cmd, Client: client}
	return cmdCtx, plugin, nil
}

type serveCli struct {
	LogConfig log.Config `embed:"" group:"Logging:"`
	Socket    string     `help:"Socket to listen on." env:"FTL_PLUGIN_SOCKET" required:""`
	kong.Plugins
}

// Start a gRPC server plugin listening on the socket specified by the
// environment variable FTL_PLUGIN_SOCKET.
//
// "create" is called to create the implementation of the service.
// "Config" is Kong configuration to pass to "create".
// "register" is called to register the service with the gRPC server.
func Start[Impl any, Iface any, Config any](
	create func(context.Context, Config) (Impl, error),
	register func(grpc.ServiceRegistrar, Iface),
) {
	var config Config
	cli := serveCli{Plugins: kong.Plugins{&config}}
	kctx := kong.Parse(&cli, kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`))

	name := "FTL." + kctx.Model.Name

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := log.New(cli.LogConfig, os.Stderr).With("C", name)
	ctx = log.ContextWithLogger(ctx, logger)

	logger.Info("Starting "+name, "socket", cli.Socket)

	// Signal handling.
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigch
		logger.Info(name+" terminating", "signal", sig)
		cancel()
		_ = syscall.Kill(-syscall.Getpid(), sig.(syscall.Signal)) //nolint:forcetypeassert
		os.Exit(0)
	}()

	svc, err := create(ctx, config)
	kctx.FatalIfErrorf(err)

	_ = os.Remove(cli.Socket)
	l, err := (&net.ListenConfig{}).Listen(ctx, "unix", cli.Socket)
	kctx.FatalIfErrorf(err)
	gs := grpc.NewServer(
		grpc.ChainUnaryInterceptor(log.UnaryGRPCInterceptor(logger)),
		grpc.ChainStreamInterceptor(log.StreamGRPCInterceptor(logger)),
	)
	reflection.Register(gs)
	register(gs, any(svc).(Iface)) //nolint:forcetypeassert
	err = gs.Serve(l)
	kctx.FatalIfErrorf(err)
}
