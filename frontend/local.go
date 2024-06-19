//go:build !release

package frontend

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/cors"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

var proxyURL, _ = url.Parse("http://localhost:5173") //nolint:errcheck
var proxy = httputil.NewSingleHostReverseProxy(proxyURL)

func Server(ctx context.Context, timestamp time.Time, publicURL *url.URL, allowOrigin *url.URL) (http.Handler, error) {
	logger := log.FromContext(ctx)
	logger.Debugf("Building console...")

	gitRoot, ok := internal.GitRoot("").Get()
	if !ok {
		return nil, fmt.Errorf("failed to find Git root")
	}

	err := exec.Command(ctx, log.Debug, gitRoot, "just", "build-frontend").RunBuffered(ctx)
	if err != nil {
		return nil, err
	}

	err = exec.Command(ctx, log.Debug, "frontend", "npm", "run", "dev").Start()
	if err != nil {
		return nil, err
	}

	if allowOrigin == nil {
		return proxy, nil
	}

	return cors.Middleware([]string{allowOrigin.String()}, proxy), nil
}
