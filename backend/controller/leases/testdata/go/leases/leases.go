package leases

import (
	"context"
	"fmt"
	"time"

	"github.com/block/ftl/go-runtime/ftl"
)

// Acquire acquires a lease and waits 5s before releasing it.
//
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
