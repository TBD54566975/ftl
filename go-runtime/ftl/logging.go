package ftl

import (
	"context"

	"github.com/TBD54566975/ftl/backend/common/log"
)

// Logger is a levelled printf-style logger with support for structured
// attributes.
type Logger = log.Logger

// Log levels.
const (
	Trace   = log.Trace
	Debug   = log.Debug
	Info    = log.Info
	Warn    = log.Warn
	Error   = log.Error
	Default = log.Default
)

// LoggerFromContext retrieves the current logger from the Context.
func LoggerFromContext(ctx context.Context) *Logger {
	return log.FromContext(ctx)
}
