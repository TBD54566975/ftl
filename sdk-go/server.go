package sdkgo

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"syscall"
	"time"

	"github.com/alecthomas/errors"
	"github.com/cloudflare/tableflip"

	"github.com/TBD54566975/ftl/common/log"
)

// Serve starts a hot swappable HTTP server.
func Serve(ctx context.Context, handler http.Handler) {
	upg, err := tableflip.New(tableflip.Options{})
	if err != nil {
		panic(err)
	}
	defer upg.Stop()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP)
		for range sig {
			err = upg.Upgrade()
			if err != nil {
				panic(err)
			}
		}
	}()

	l, err := upg.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	defer l.Close() //nolint:gosec

	srv := http.Server{
		Handler:     handler,
		Addr:        "127.0.0.1:8080",
		BaseContext: func(_ net.Listener) context.Context { return ctx },
		ReadTimeout: time.Minute * 5,
	}

	go func() {
		err := srv.Serve(l)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	if err := upg.Ready(); err != nil {
		panic(err)
	}
	<-upg.Exit()
}

// Handler converts a Verb function into a http.Handler.
func Handler[Req, Resp any](verb func(ctx context.Context, req Req) (Resp, error)) http.Handler {
	name := runtime.FuncForPC(reflect.ValueOf(verb).Pointer()).Name()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.FromContext(r.Context())
		caller := r.Header.Get("X-FTL-Caller")
		if caller == "" {
			caller = "?"
		}
		logger.Info("Call", "s", caller, "d", name, "ua", r.Header.Get("User-Agent"))

		// Decode request.
		var req Req
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid request to verb %s: %s", name, err.Error()), http.StatusBadRequest)
			return
		}

		// Call Verb.
		resp, err := verb(r.Context(), req)
		if err != nil {
			http.Error(w, fmt.Sprintf("Call to Verb %s failed: %s", name, err.Error()), http.StatusInternalServerError)
			return
		}

		// Encode response.
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		err = enc.Encode(resp)
		if err != nil {
			logger.Error("Failed to encode response", err)
		}
	})
}
