//go:build !release

package console

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

var consoleURL, _ = url.Parse("http://localhost:5173")
var proxy = httputil.NewSingleHostReverseProxy(consoleURL)

func Server(ctx context.Context) (http.Handler, error) {
	logger := log.FromContext(ctx)
	logger.Infof("Building console...")

	err := exec.Command(ctx, "console/client", "npm", "install").Run()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = exec.Command(ctx, "console/client", "npm", "run", "dev").Start()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	logger.Infof("Console started")

	return http.HandlerFunc(handler), nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	proxy.ServeHTTP(w, r)
}
