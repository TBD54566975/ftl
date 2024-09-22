package log

import (
	"fmt"
	"io"
	"os"
	"time"

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

func newPlainSink(w io.Writer, logTime bool, alwaysColor bool) *plainSink {
	var isaTTY bool
	if alwaysColor {
		isaTTY = true
	} else if f, ok := w.(*os.File); ok {
		isaTTY = isatty.IsTerminal(f.Fd())
	}
	return &plainSink{
		isaTTY:  isaTTY,
		w:       w,
		logTime: logTime,
	}
}

type plainSink struct {
	isaTTY  bool
	w       io.Writer
	logTime bool
}

// Log implements Sink
func (t *plainSink) Log(entry Entry) error {
	var prefix string

	// Add timestamp if required
	if t.logTime {
		prefix += entry.Time.Format(time.TimeOnly) + " "
	}

	// Add scope if required
	scope, exists := entry.Attributes[scopeKey]
	if exists {
		if t.isaTTY {
			prefix += entry.Level.String() + ":\u001B[35m" + scope + "\u001B[0m: "
		} else {
			prefix += entry.Level.String() + ":" + scope + ": "
		}
	} else {
		prefix += entry.Level.String() + ": "
	}

	// Print
	var err error
	if t.isaTTY {
		_, err = fmt.Fprintf(t.w, "%s%s%s%s\x1b[0m\n", colours[entry.Level], prefix, colours[entry.Level], entry.Message)
	} else {
		_, err = fmt.Fprintf(t.w, "%s%s\n", prefix, entry.Message)
	}
	if err != nil {
		return err
	}
	return nil
}
