package main

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/alecthomas/errors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	ftlv1 "github.com/TBD54566975/ftl/common/gen/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/common/socket"
	"github.com/TBD54566975/ftl/local"
)

type serveCmd struct {
	Dir []string `arg:"" help:"Paths to FTL modules."`
}

func (r *serveCmd) Run(ctx context.Context, s socket.Socket) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg, ctx := errgroup.WithContext(ctx)

	l, err := socket.Listen(s)
	if err != nil {
		return errors.WithStack(err)
	}

	agent, err := local.New(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	// Spawn modules in parallel.
	spawnwg := errgroup.Group{}
	for _, dir := range r.Dir {
		dir := dir
		spawnwg.Go(func() error { return agent.Manage(ctx, dir) })
	}
	err = spawnwg.Wait()
	if err != nil {
		return errors.WithStack(err)
	}

	logger := log.FromContext(ctx)

	// Start agent gRPC and REST servers.
	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(log.UnaryGRPCInterceptor(logger)),
		grpc.ChainStreamInterceptor(log.StreamGRPCInterceptor(logger)),
	)
	ftlv1.RegisterAgentServiceServer(srv, agent)

	mixedHandler := newHTTPandGRPCMux(agent, srv)
	http2Server := &http2.Server{}
	http1Server := &http.Server{Handler: h2c.NewHandler(mixedHandler, http2Server), ReadHeaderTimeout: time.Second * 30}

	wg.Go(func() error { return errors.WithStack(http1Server.Serve(l)) })
	wg.Go(agent.Wait)
	return errors.WithStack(wg.Wait())
}

func newHTTPandGRPCMux(httpHand http.Handler, grpcHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("content-type"), "application/grpc") {
			grpcHandler.ServeHTTP(w, r)
			return
		}
		httpHand.ServeHTTP(w, r)
	})
}
