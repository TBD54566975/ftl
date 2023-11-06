//go:build !release

package frontend

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/alecthomas/errors"

	"github.com/TBD54566975/ftl/backend/common/cors"
	"github.com/TBD54566975/ftl/backend/common/exec"
	"github.com/TBD54566975/ftl/backend/common/log"
)

var consoleURL, _ = url.Parse("http://localhost:5173")
var proxy = httputil.NewSingleHostReverseProxy(consoleURL)

func Server(ctx context.Context, timestamp time.Time, allowOrigin *url.URL) (http.Handler, error) {
	logger := log.FromContext(ctx)
	logger.Infof("Building console...")

	err := exec.Command(ctx, log.Debug, "frontend", "npm", "install").Run()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = exec.Command(ctx, log.Debug, "frontend", "npm", "run", "dev").Start()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	logger.Infof("Console started")

	if allowOrigin == nil {
		return proxy, nil
	}

	return cors.Middleware([]string{allowOrigin.String()}, proxy), nil
}
