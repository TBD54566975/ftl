package rpc

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/alecthomas/concurrency"
	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/TBD54566975/ftl/internal/pubsub"
)

const ShutdownGracePeriod = time.Second * 5

type optionsBundle struct {
	mux *http.ServeMux
}

type Option func(*optionsBundle)

type GRPCServerConstructor[Iface Pingable] func(svc Iface, opts ...connect.HandlerOption) (string, http.Handler)

// GRPC is a convenience function for registering a GRPC server with default options.
func GRPC[Iface, Impl Pingable](constructor GRPCServerConstructor[Iface], impl Impl, options ...connect.HandlerOption) Option {
	return func(o *optionsBundle) {
		options = append(options, DefaultHandlerOptions()...)
		path, handler := constructor(any(impl).(Iface), options...)
		o.mux.Handle(path, handler)
	}
}

func Route(prefix string, handler http.Handler) Option {
	return func(o *optionsBundle) {
		o.mux.Handle(prefix, handler)
	}
}

type Server struct {
	listen *url.URL
	Bind   *pubsub.Topic[*url.URL] // Will be updated with the actual bind address.
	Server *http.Server
}

func NewServer(ctx context.Context, listen *url.URL, options ...Option) (*Server, error) {
	opts := &optionsBundle{
		mux: http.NewServeMux(),
	}

	for _, option := range options {
		option(opts)
	}
	// TODO: Is this a good idea? Who knows!
	crs := cors.New(cors.Options{
		AllowedOrigins: []string{listen.String()},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
	})
	root := crs.Handler(ContextValuesMiddleware(ctx, opts.mux))

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
