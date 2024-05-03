package leases

import (
	"context"
	"fmt"
	"time"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

//ftl:verb
func Acquire(ctx context.Context) error {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("Acquiring lease")
	lease, err := ftl.Lease(ctx, 10*time.Second, "lease")
	if err != nil {
		logger.Warnf("Failed to acquire lease: %s", err)
		return fmt.Errorf("failed to acquire lease: %w", err)
	}
	logger.Infof("Acquired lease!")
	time.Sleep(time.Second * 5)
	return lease.Release()
}
