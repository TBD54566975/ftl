// Package server contains code common to all servers in FTL.
package server

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/TBD54566975/ftl/common/rpc"
	"github.com/TBD54566975/ftl/common/socket"
)

type optionsBundle struct {
	mux *http.ServeMux
}

type Option func(*optionsBundle)

type GRPCServerConstructor[Iface rpc.Pingable] func(svc Iface, opts ...connect.HandlerOption) (string, http.Handler)

// GRPC is a convenience function for registering a GRPC server with default options.
func GRPC[Iface, Impl rpc.Pingable](constructor GRPCServerConstructor[Iface], impl Impl, options ...connect.HandlerOption) Option {
	return func(o *optionsBundle) {
		options = append(options, rpc.DefaultHandlerOptions()...)
		path, handler := constructor(any(impl).(Iface), options...)
		o.mux.Handle(path, handler)
	}
}

func Route(prefix string, handler http.Handler) Option {
	return func(o *optionsBundle) {
		o.mux.Handle(prefix, handler)
	}
}

func Serve(ctx context.Context, listen socket.Socket, options ...Option) error {
	opts := &optionsBundle{
		mux: http.NewServeMux(),
	}

	for _, option := range options {
		option(opts)
	}
	// TODO: Is this a good idea? Who knows!
	crs := cors.New(cors.Options{
		AllowedOrigins: []string{listen.URL().String()},
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
	root := crs.Handler(rpc.Middleware(ctx, opts.mux))

	http1Server := &http.Server{
		Handler:           h2c.NewHandler(root, &http2.Server{}),
		ReadHeaderTimeout: time.Second * 30,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

	listener, err := socket.Listen(listen)
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(http1Server.Serve(listener))
}
