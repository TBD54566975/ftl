package log

import (
	"fmt"
	"os"
)

var _ Interface = (*Logger)(nil)

type Entry struct {
	Level   Level    `json:"level"`
	Scope   []string `json:"scope,omitempty"`
	Message string   `json:"message"`

	Error error `json:"error,omitempty"`
}

// Logger is the concrete logger.
type Logger struct {
	level Level
	scope []string
	sink  Sink
}

// New returns a new logger.
func New(level Level, sink Sink) *Logger {
	return &Logger{
		level: level,
		sink:  sink,
	}
}

// Sub creates a new logger with the given prefix.
func (l Logger) Sub(scope string, level Level) *Logger {
	if scope != "" {
		l.scope = append(l.scope, scope)
	}
	if level != Default {
		l.level = level
	}
	return &l
}

func (l *Logger) Level() Level {
	return l.level
}

func (l *Logger) Log(entry Entry) {
	if entry.Level < l.level {
		return
	}
	entry.Scope = l.scope
	if err := l.sink.Log(entry); err != nil {
		fmt.Fprintf(os.Stderr, "ftl:log: failed to log entry: %v", err)
	}
}

func (l *Logger) Tracef(format string, args ...interface{}) {
	l.Log(Entry{Level: Trace, Message: fmt.Sprintf(format, args...)})
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Log(Entry{Level: Debug, Message: fmt.Sprintf(format, args...)})
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.Log(Entry{Level: Info, Message: fmt.Sprintf(format, args...)})
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Log(Entry{Level: Warn, Message: fmt.Sprintf(format, args...)})
}

func (l *Logger) Errorf(err error, format string, args ...interface{}) {
	l.Log(Entry{Level: Error, Message: fmt.Sprintf(format, args...), Error: err})
}
