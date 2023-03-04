package log

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
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
		handler = &cliHandler{w: w, isaTTY: isatty.IsTerminal(os.Stdout.Fd()), level: config.Level}
	}
	return slog.New(handler)
}

var colours map[slog.Level]string = map[slog.Level]string{
	slog.LevelDebug: "\x1b[34m", // Blue
	slog.LevelInfo:  "\x1b[32m", // Green
	slog.LevelWarn:  "\x1b[33m", // Yellow
	slog.LevelError: "\x1b[31m", // Red
}

type cliHandler struct {
	parent *cliHandler
	group  string
	attrs  []slog.Attr
	w      io.Writer
	level  slog.Level
	isaTTY bool
}

func (c *cliHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= c.level
}

func (c *cliHandler) Handle(ctx context.Context, record slog.Record) error {
	if c.isaTTY {
		fmt.Fprint(c.w, colours[record.Level]+"\x1b[1m")
	}
	fmt.Fprintf(c.w, "%s", record.Message)
	if c.isaTTY {
		fmt.Fprintf(c.w, "\x1b[0m")
	}
	for _, a := range c.allAttrs(record) {
		c.printAttr(a)
	}
	if c.isaTTY {
		fmt.Fprint(c.w, "\x1b[0m")
	}
	fmt.Fprintln(c.w)
	return nil
}

func (c *cliHandler) allAttrs(record slog.Record) (attrs []slog.Attr) {
	record.Attrs(func(a slog.Attr) {
		attrs = append(attrs, a)
	})
	for p := c; p != nil; p = p.parent {
		attrs = append(attrs, p.attrs...)
	}
	return attrs
}

func (c *cliHandler) printAttr(a slog.Attr) {
	fmt.Fprint(c.w, " ")
	if c.isaTTY {
		fmt.Fprint(c.w, "\x1b[0m")
	}
	fmt.Fprintf(c.w, "%s", a.Key)
	if c.isaTTY {
		fmt.Fprintf(c.w, "\x1b[0m\x1b[2m")
	}
	fmt.Fprintf(c.w, "=%v", a.Value)
}

func (c *cliHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := *c
	out.parent = c
	out.attrs = attrs
	return &out
}

func (c *cliHandler) WithGroup(name string) slog.Handler {
	out := *c
	out.parent = c
	out.group = name
	return &out
}
