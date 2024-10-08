package server

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/TBD54566975/ftl/internal/log"
)

// RunWithSignalHandler installs a signal handler that cancels the context on SIGINT and SIGTERM, sends a SIGTERM to the
// process group, and exits.
func RunWithSignalHandler(ctx context.Context, run func(ctx context.Context) error) error {
	logger := log.FromContext(ctx)
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	err := run(ctx)
	logger.Debugf("Shutting down")
	_ = syscall.Kill(-syscall.Getpid(), syscall.SIGTERM) //nolint:errcheck // best effort
	if err == nil {
		return nil
	} else if errors.Is(err, context.Canceled) {
		return nil
	}
	return fmt.Errorf("%w", err)
}
