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

func Server(ctx context.Context) (http.Handler, error) {
	logger := log.FromContext(ctx)
	logger.Infof("Starting console dev server")
	output := logger.WriterAt(log.Info)
	cmd := exec.Command(ctx, "console", "npm", "run", "dev")
	cmd.Stdout = output
	cmd.Stderr = output
	err := cmd.Start()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   "localhost:5173",
	}), nil
}
