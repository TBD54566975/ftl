//go:build !release

package console

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/TBD54566975/ftl/internal"
	"github.com/TBD54566975/ftl/internal/cors"
	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/flock"
	"github.com/TBD54566975/ftl/internal/log"
)

var proxyURL, _ = url.Parse("http://localhost:5173") //nolint:errcheck
var proxy = httputil.NewSingleHostReverseProxy(proxyURL)

func Server(ctx context.Context, timestamp time.Time, publicURL *url.URL, allowOrigin *url.URL) (http.Handler, error) {
	gitRoot, ok := internal.GitRoot(os.Getenv("FTL_DIR")).Get()
	if !ok {
		return nil, fmt.Errorf("failed to find Git root")
	}

	// Lock the frontend directory to prevent concurrent builds.
	release, err := flock.Acquire(ctx, filepath.Join(gitRoot, ".frontend.lock"), 2*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	logger := log.FromContext(ctx)
	logger.Debugf("Building console...")

	err = exec.Command(ctx, log.Debug, gitRoot, "just", "build-frontend").RunBuffered(ctx)
	if lerr := release(); lerr != nil {
		return nil, errors.Join(fmt.Errorf("failed to release lock: %w", lerr))
	}
	if err != nil {
		return nil, err
	}

	err = exec.Command(ctx, log.Debug, path.Join(gitRoot, "frontend", "console"), "pnpm", "run", "dev").Start()
	if err != nil {
		return nil, err
	}

	if allowOrigin == nil {
		return proxy, nil
	}

	return cors.Middleware([]string{allowOrigin.String()}, nil, proxy), nil
}
