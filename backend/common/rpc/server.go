package rpc

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/alecthomas/concurrency"
	"github.com/alecthomas/errors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/TBD54566975/ftl/backend/common/pubsub"
)

const ShutdownGracePeriod = time.Second * 5

type serverOptions struct {
	mux             *http.ServeMux
	reflectionPaths []string
}

type Option func(*serverOptions)

type GRPCServerConstructor[Iface Pingable] func(svc Iface, opts ...connect.HandlerOption) (string, http.Handler)
type RawGRPCServerConstructor[Iface any] func(svc Iface, opts ...connect.HandlerOption) (string, http.Handler)

// GRPC is a convenience function for registering a GRPC server with default options.
// TODO(aat): Do we need pingable here?
func GRPC[Iface, Impl Pingable](constructor GRPCServerConstructor[Iface], impl Impl, options ...connect.HandlerOption) Option {
	return func(o *serverOptions) {
		options = append(options, DefaultHandlerOptions()...)
		path, handler := constructor(any(impl).(Iface), options...)
		o.reflectionPaths = append(o.reflectionPaths, strings.Trim(path, "/"))
		o.mux.Handle(path, handler)
	}
}

// RawGRPC is a convenience function for registering a GRPC server with default options without Pingable.
func RawGRPC[Iface, Impl any](constructor RawGRPCServerConstructor[Iface], impl Impl, options ...connect.HandlerOption) Option {
	return func(o *serverOptions) {
		options = append(options, DefaultHandlerOptions()...)
		path, handler := constructor(any(impl).(Iface), options...)
		o.reflectionPaths = append(o.reflectionPaths, strings.Trim(path, "/"))
		o.mux.Handle(path, handler)
	}
}

// HTTP adds a HTTP route to the server.
func HTTP(prefix string, handler http.Handler) Option {
	return func(o *serverOptions) {
		o.mux.Handle(prefix, handler)
	}
}

type Server struct {
	listen *url.URL
	Bind   *pubsub.Topic[*url.URL] // Will be updated with the actual bind address.
	Server *http.Server
}

func NewServer(ctx context.Context, listen *url.URL, options ...Option) (*Server, error) {
	opts := &serverOptions{
		mux: http.NewServeMux(),
	}

	opts.mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for _, option := range options {
		option(opts)
	}

	// Register reflection services.
	reflector := grpcreflect.NewStaticReflector(opts.reflectionPaths...)
	opts.mux.Handle(grpcreflect.NewHandlerV1(reflector))
	opts.mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	root := ContextValuesMiddleware(ctx, opts.mux)

	http1Server := &http.Server{
		Handler:           h2c.NewHandler(root, &http2.Server{}),
		ReadHeaderTimeout: time.Second * 30,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

	return &Server{
		listen: listen,
		Bind:   pubsub.New[*url.URL](),
		Server: http1Server,
	}, nil
}

// Serve runs the server, updating .Bind with the actual bind address.
func (s *Server) Serve(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.listen.Host)
	if err != nil {
		return errors.WithStack(err)
	}
	if s.listen.Port() == "0" {
		s.listen.Host = listener.Addr().String()
	}
	s.Bind.Publish(s.listen)

	tree, _ := concurrency.New(ctx)

	// Shutdown server on context cancellation.
	tree.Go(func(ctx context.Context) error {
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), ShutdownGracePeriod)
		defer cancel()
		err := s.Server.Shutdown(ctx)
		if err == nil {
			return nil
		}
		if errors.Is(err, context.Canceled) {
			_ = s.Server.Close()
			return errors.WithStack(err)
		}
		return errors.WithStack(err)
	})

	// Start server.
	tree.Go(func(ctx context.Context) error {
		err = s.Server.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return errors.WithStack(err)
	})

	return errors.WithStack(tree.Wait())
}

// Serve starts a HTTP and Connect gRPC server with sane defaults for FTL.
//
// Blocks until the context is cancelled.
func Serve(ctx context.Context, listen *url.URL, options ...Option) error {
	server, err := NewServer(ctx, listen, options...)
	if err != nil {
		return errors.WithStack(err)
	}
	return server.Serve(ctx)
}
