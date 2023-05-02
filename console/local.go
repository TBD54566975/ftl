//go:build !release

package console

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/common/exec"
	"github.com/TBD54566975/ftl/common/log"
)

var consoleURL, _ = url.Parse("http://localhost:5173")
var proxy = httputil.NewSingleHostReverseProxy(consoleURL)

func Server(ctx context.Context) (http.Handler, error) {
	logger := log.FromContext(ctx)
	logger.Infof("Starting console dev server")

	err := exec.Command(ctx, "console", "npm", "install").Run()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = exec.Command(ctx, "console", "npm", "run", "dev").Start()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return http.HandlerFunc(handler), nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	proxy.ServeHTTP(w, r)
}
