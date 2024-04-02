package lsp

import (
	"fmt"

	"github.com/tliron/commonlog"

	"github.com/TBD54566975/ftl/internal/log"
)

// GLSPLogger is a custom logger for the language server.
type GLSPLogger struct {
	commonlog.Logger
}

func (l *GLSPLogger) Log(entry log.Entry) {
	l.Logger.Log(toGLSPLevel(entry.Level), 10, entry.Message, entry.Attributes)
}

func (l *GLSPLogger) Logf(level log.Level, format string, args ...interface{}) {
	l.Log(log.Entry{Level: level, Message: fmt.Sprintf(format, args...)})
}
func (l *GLSPLogger) Tracef(format string, args ...interface{}) {
	l.Log(log.Entry{Level: log.Trace, Message: fmt.Sprintf(format, args...)})
}

func (l *GLSPLogger) Debugf(format string, args ...interface{}) {
	l.Log(log.Entry{Level: log.Debug, Message: fmt.Sprintf(format, args...)})
}

func (l *GLSPLogger) Infof(format string, args ...interface{}) {
	l.Log(log.Entry{Level: log.Info, Message: fmt.Sprintf(format, args...)})
}

func (l *GLSPLogger) Warnf(format string, args ...interface{}) {
	l.Log(log.Entry{Level: log.Warn, Message: fmt.Sprintf(format, args...)})
}

func (l *GLSPLogger) Errorf(err error, format string, args ...interface{}) {
	if err == nil {
		return
	}
	l.Log(log.Entry{Level: log.Error, Message: fmt.Sprintf(format, args...) + ": " + err.Error(), Error: err})
}

var _ log.Interface = (*GLSPLogger)(nil)

func NewGLSPLogger(log commonlog.Logger) *GLSPLogger {
	return &GLSPLogger{log}
}

func toGLSPLevel(l log.Level) commonlog.Level {
	switch l {
	case log.Trace:
		return commonlog.Debug
	case log.Debug:
		return commonlog.Debug
	case log.Info:
		return commonlog.Info
	case log.Warn:
		return commonlog.Warning
	case log.Error:
		return commonlog.Error
	default:
		return commonlog.Debug
	}
}
