package log

import (
	"context"
	"io"

	"golang.org/x/exp/slog"
)

var contextKey = struct{ string }{"logger"}

// Config for the a logger.
type Config struct {
	Level      slog.Level `help:"Log level." default:"info"`
	WithSource bool       `help:"Include source locations."`
	JSON       bool       `help:"Log in JSON format."`
}

// FromContext retrieves the current logger from the context or panics
func FromContext(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(contextKey).(*slog.Logger)
	if ok {
		return logger
	}
	panic("no logger in context")
}

// ContextWithLogger returns a new context with the given logger attached. Use
// FromContext to retrieve it.
func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, contextKey, logger)
}

// New returns a new logger.
//
// This should typically not be used, instead the root logger should be
// extracted from the context.
func New(config Config, w io.Writer) *slog.Logger {
	loptions := slog.HandlerOptions{
		Level:     config.Level,
		AddSource: config.WithSource,
	}
	var handler slog.Handler
	if config.JSON {
		handler = loptions.NewJSONHandler(w)
	} else {
		loptions.ReplaceAttr = func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "time" {
				return slog.Attr{}
			}
			return a
		}
		handler = loptions.NewTextHandler(w)
	}
	return slog.New(handler)
}
