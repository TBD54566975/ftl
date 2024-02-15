// Package ftltest contains test utilities for the ftl package.
package ftltest

import (
	"context"
	"os"

	"github.com/TBD54566975/ftl/backend/common/log"
)

// Context suitable for use in testing FTL verbs.
func Context() context.Context {
	return log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, log.Config{Level: log.Trace}))
}
