package log

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/mattn/go-isatty"
)

var colours = map[Level]string{
	Trace: "\x1b[90m", // Dark gray
	Debug: "\x1b[34m", // Blue
	Info:  "\x1b[37m", // White
	Warn:  "\x1b[33m", // Yellow
	Error: "\x1b[31m", // Red
}

var _ Sink = (*plainSink)(nil)

func newPlainSink(w io.Writer) *plainSink {
	var isaTTY bool
	if f, ok := w.(*os.File); ok {
		isaTTY = isatty.IsTerminal(f.Fd())
	}
	return &plainSink{
		isaTTY: isaTTY,
		w:      w,
	}
}

type plainSink struct {
	isaTTY bool
	w      io.Writer
}

// Log implements Sink
func (t *plainSink) Log(entry Entry) error {
	var prefix string
	if len(entry.Scope) > 0 {
		prefix = entry.Level.String() + ":" + strings.Join(entry.Scope, ":") + ": "
	} else {
		prefix = entry.Level.String() + ": "
	}
	var err error
	if t.isaTTY {
		_, err = fmt.Fprintf(t.w, "%s%s%s\x1b[0m\n", colours[entry.Level], prefix, entry.Message)
	} else {
		_, err = fmt.Fprintf(t.w, "%s%s\n", prefix, entry.Message)
	}
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
