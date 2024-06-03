package http

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/TBD54566975/ftl/internal/log"
)

const ShutdownGracePeriod = 5 * time.Second

func Serve(ctx context.Context, listen *url.URL, handler http.Handler) error {
	httpServer := &http.Server{
		Addr:              listen.Host,
		Handler:           handler,
		ReadHeaderTimeout: 30 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownGracePeriod)
		defer cancel()
		err := httpServer.Shutdown(shutdownCtx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				_ = httpServer.Close()
				return
			}
			log.FromContext(ctx).Errorf(err, "server shutdown error")
		}
	}()

	err := httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
