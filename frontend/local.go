//go:build !release

package frontend

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/cors"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

var proxyURL, _ = url.Parse("http://localhost:5173")
var proxy = httputil.NewSingleHostReverseProxy(proxyURL)

func Server(ctx context.Context, timestamp time.Time, publicURL *url.URL, allowOrigin *url.URL) (http.Handler, error) {
	logger := log.FromContext(ctx)
	logger.Debugf("Building console...")

	err := exec.Command(ctx, log.Debug, internal.GitRoot(""), "bit", "frontend/**/*").RunBuffered(ctx)
	if err != nil {
		return nil, err
	}

	err = exec.Command(ctx, log.Debug, "frontend", "npm", "run", "dev").Start()
	if err != nil {
		return nil, err
	}
	logger.Infof("Web console available at: %s", publicURL.String())

	if allowOrigin == nil {
		return proxy, nil
	}

	return cors.Middleware([]string{allowOrigin.String()}, proxy), nil
}
