package cron

import (
	"context"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

//ftl:cron 30s
func ThirtySeconds(ctx context.Context) error {
	ftl.LoggerFromContext(ctx).Infof("Frequent cron job triggered.")
	return nil
}

//ftl:cron 0 * * * *
func Hourly(ctx context.Context) error {
	ftl.LoggerFromContext(ctx).Infof("Hourly cron job triggered.")
	return nil
}
