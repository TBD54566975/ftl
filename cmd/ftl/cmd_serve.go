package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/errors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/reflection"

	"github.com/TBD54566975/ftl/agent"
	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/common/socket"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
)

type serveCmd struct {
	Dir []string `arg:"" help:"Paths to FTL modules."`
}

func (r *serveCmd) Run(ctx context.Context, s socket.Socket) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger := log.FromContext(ctx).Sub("agent", log.Default)
	logger.Infof("Starting FTL local agent")
	ctx = log.ContextWithLogger(ctx, logger)

	wg, ctx := errgroup.WithContext(ctx)

	l, err := socket.Listen(s)
	if err != nil {
		return errors.WithStack(err)
	}
	// Used by sub-processes to call back into FTL.
	os.Setenv("FTL_ENDPOINT", s.String())

	agent, err := agent.New(ctx)
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

	// Start agent gRPC and REST servers.
	srv := socket.NewGRPCServer(ctx)
	reflection.Register(srv)
	ftlv1.RegisterVerbServiceServer(srv, agent)
	ftlv1.RegisterDevelServiceServer(srv, agent)

	mixedHandler := newHTTPandGRPCMux(agent, srv)
	http2Server := &http2.Server{}
	http1Server := &http.Server{
		Handler:           h2c.NewHandler(mixedHandler, http2Server),
		ReadHeaderTimeout: time.Second * 30,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

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
