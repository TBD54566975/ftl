package cron

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
)

//ftl:cron 2s
func Job(ctx context.Context) error {
	ftl.LoggerFromContext(ctx).Infof("Frequent cron job triggered.")
	return nil
}
