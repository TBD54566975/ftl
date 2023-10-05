package plugin

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"syscall"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/alecthomas/errors"
	"github.com/alecthomas/kong"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	_ "github.com/TBD54566975/ftl/backend/common/automaxprocs" // Set GOMAXPROCS to match Linux container CPU quota.
	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/rpc"
)

type serveCli struct {
	LogConfig log.Config `prefix:"log-" embed:"" group:"Logging:"`
	Bind      *url.URL   `help:"URL to listen on." env:"FTL_BIND" required:""`
	kong.Plugins
}

type serverRegister[Impl any] struct {
	servicePath string
	register    func(i Impl, mux *http.ServeMux)
}

type handlerPath struct {
	path    string
	handler http.Handler
}

type startOptions[Impl any] struct {
	register []serverRegister[Impl]
	handlers []handlerPath
}

// StartOption is an option for Start.
type StartOption[Impl any] func(*startOptions[Impl])

// ConnectHandlerFactory is a type alias for a function that creates a new Connect request handler.
//
// This will typically just be the generated NewXYZHandler function.
type ConnectHandlerFactory[Iface any] func(Iface, ...connect.HandlerOption) (string, http.Handler)

// RegisterAdditionalServer allows a plugin to serve additional gRPC services.
//
// "Impl" must be an implementation of "Iface.
func RegisterAdditionalServer[Impl any, Iface any](servicePath string, register ConnectHandlerFactory[Iface]) StartOption[Impl] {
	return func(so *startOptions[Impl]) {
		so.register = append(so.register, serverRegister[Impl]{
			servicePath: servicePath,
			register: func(i Impl, mux *http.ServeMux) {
				mux.Handle(register(any(i).(Iface), rpc.DefaultHandlerOptions()...)) //nolint:forcetypeassert
			}})
	}
}

// RegisterAdditionalHandler allows a plugin to serve additional HTTP handlers.
func RegisterAdditionalHandler[Impl any](path string, handler http.Handler) StartOption[Impl] {
	return func(so *startOptions[Impl]) {
		so.handlers = append(so.handlers, handlerPath{path: path, handler: handler})
	}
}

// Constructor is a function that creates a new plugin server implementation.
type Constructor[Impl any, Config any] func(context.Context, Config) (context.Context, Impl, error)

// Start a gRPC server plugin listening on the socket specified by the
// environment variable FTL_BIND.
//
// This function does not return.
//
// "Config" is Kong configuration to pass to "create".
// "create" is called to create the implementation of the service.
// "register" is called to register the service with the gRPC server and is typically a generated function.
func Start[Impl any, Iface any, Config any](
	ctx context.Context,
	name string,
	create Constructor[Impl, Config],
	servicePath string,
	register ConnectHandlerFactory[Iface],
	options ...StartOption[Impl],
) {
	var config Config
	cli := serveCli{Plugins: kong.Plugins{&config}}
	kctx := kong.Parse(&cli, kong.Description(`FTL - Towards a 𝝺-calculus for large-scale systems`))

	mux := http.NewServeMux()

	so := &startOptions[Impl]{}
	for _, option := range options {
		option(so)
	}

	for _, handler := range so.handlers {
		mux.Handle(handler.path, handler.handler)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Configure logging to JSON on stdout. This will be read by the parent process.
	logConfig := cli.LogConfig
	logConfig.JSON = true
	logger := log.Configure(os.Stderr, logConfig)

	logger = logger.Scope(name)
	ctx = log.ContextWithLogger(ctx, logger)

	logger.Debugf("Starting on %s", cli.Bind)

	// Signal handling.
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigch
		logger.Infof("Terminated by signal %s", sig)
		cancel()
		_ = syscall.Kill(-syscall.Getpid(), sig.(syscall.Signal)) //nolint:forcetypeassert
		os.Exit(0)
	}()

	ctx, svc, err := create(ctx, config)
	kctx.FatalIfErrorf(err)

	if _, ok := any(svc).(Iface); !ok {
		var iface Iface
		panic(fmt.Sprintf("%s does not implement %s", reflect.TypeOf(svc), reflect.TypeOf(iface)))
	}

	l, err := net.Listen("tcp", cli.Bind.Host)
	kctx.FatalIfErrorf(err)

	servicePaths := []string{servicePath}

	mux.Handle(register(any(svc).(Iface), rpc.DefaultHandlerOptions()...)) //nolint:forcetypeassert
	for _, register := range so.register {
		register.register(svc, mux)
		servicePaths = append(servicePaths, register.servicePath)
	}

	reflector := grpcreflect.NewStaticReflector(servicePaths...)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// Start the server.
	http1Server := &http.Server{
		Handler:           h2c.NewHandler(mux, &http2.Server{}),
		ReadHeaderTimeout: time.Second * 30,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}
	err = http1Server.Serve(l)
	kctx.FatalIfErrorf(err)

	kctx.Exit(0)
}

func allocatePort() (*net.TCPAddr, error) {
	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		return nil, errors.Wrap(err, "failed to allocate port")
	}
	_ = l.Close()
	return l.Addr().(*net.TCPAddr), nil //nolint:forcetypeassert
}

func cleanup(logger *log.Logger, pidFile string) error {
	pidb, err := os.ReadFile(pidFile)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return errors.WithStack(err)
	}
	pid, err := strconv.Atoi(string(pidb))
	if err != nil && !os.IsNotExist(err) {
		return errors.WithStack(err)
	}
	err = syscall.Kill(pid, syscall.SIGKILL)
	if err != nil && !errors.Is(err, syscall.ESRCH) {
		logger.Warnf("Failed to reap old plugin with pid %d: %s", pid, err)
	}
	return nil
}
