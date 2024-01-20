package sdk

import (
	"context"

	"github.com/TBD54566975/ftl/backend/common/log"
)

// Logger is a levelled printf-style logger with support for structured
// attributes.
type Logger = log.Logger

// FromContext retrieves the current logger from the Context.
func FromContext(ctx context.Context) *Logger {
	return log.FromContext(ctx)
}
