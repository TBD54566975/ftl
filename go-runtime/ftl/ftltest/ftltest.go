// Package ftltest contains test utilities for the ftl package.
package ftltest

import (
	"context"

	"github.com/TBD54566975/ftl/internal/log"
)

// Context suitable for use in testing FTL verbs.
func Context() context.Context {
	return log.ContextWithNewDefaultLogger(context.Background())
}
