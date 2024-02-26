package log

import (
	"context"
	"os"
)

type Sink interface {
	Log(entry Entry) error
}

type Interface interface {
	Log(entry Entry)
	Logf(level Level, format string, args ...interface{})
	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	// Errorf conditionally logs an error. If err is nil, nothing is logged.
	Errorf(err error, format string, args ...interface{})
}

// Level is the log level.
//
//go:generate enumer -type=Level -text -transform=lower -output log_level_string.go
type Level int

// Log levels.
const (
	// Default is a special value that means the log level will use a default.
	Default Level = 0
	Trace   Level = 1
	Debug   Level = 5
	Info    Level = 9
	Warn    Level = 13
	Error   Level = 17
)

// Severity returns the open telemetry severity of the log level.
func (l Level) Severity() int {
	return int(l)
}

// ParseLevel parses a log level from text.
func ParseLevel(input string) (Level, error) {
	var level Level
	err := level.UnmarshalText([]byte(input))
	return level, err
}

type contextKey struct{}

// FromContext retrieves the current logger from the context or panics
func FromContext(ctx context.Context) *Logger {
	logger, ok := ctx.Value(contextKey{}).(*Logger)
	if ok {
		return logger
	}
	panic("no logger in context")
}

// ContextWithLogger returns a new context with the given logger attached. Use
// FromContext to retrieve it.
func ContextWithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, logger)
}

func ContextWithNewDefaultLogger(ctx context.Context) context.Context {
	return ContextWithLogger(ctx, Configure(os.Stderr, Config{Level: Debug}))
}
