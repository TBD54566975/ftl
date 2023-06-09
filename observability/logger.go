package observability

import (
	"context"

	"github.com/TBD54566975/ftl/internal/log"
	"github.com/go-logr/logr"
)

type logSink struct {
	name   string
	logger *log.Logger
}

var _ logr.LogSink = &logSink{}

func (l *logSink) Init(info logr.RuntimeInfo) {
}

func (l logSink) Enabled(level int) bool {
	return true
}

// otel uses the following mapping for level to our log.Level
// 4 = Info
// 8 = Debug
// 1 = Warning
func (l logSink) Info(level int, msg string, kvs ...interface{}) {
	var logLevel log.Level
	switch level {
	case 4:
		logLevel = log.Info
	case 8:
		logLevel = log.Debug
	case 1:
		logLevel = log.Warn
	default:
		logLevel = log.Trace
	}

	l.logger.Logf(logLevel, "%s: %s", l.name, msg)
	for i := 0; i < len(kvs); i += 2 {
		l.logger.Logf(logLevel, "%s: %+v  ", kvs[i], kvs[i+1])
	}
}

func (l logSink) Error(err error, msg string, kvs ...interface{}) {
	l.logger.Errorf(err, msg, kvs...)
}

func (l logSink) WithName(name string) logr.LogSink {
	return &logSink{
		name:   l.name + "." + name,
		logger: l.logger,
	}
}

func (l logSink) WithValues(kvs ...interface{}) logr.LogSink {
	return &logSink{
		name:   l.name,
		logger: l.logger,
	}
}

func NewOtelLogger(ctx context.Context) logr.Logger {
	logger := log.FromContext(ctx)

	sink := &logSink{
		name:   "otel",
		logger: logger,
	}
	return logr.New(sink)
}
