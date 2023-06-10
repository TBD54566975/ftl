package observability

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	"github.com/TBD54566975/ftl/internal/log"
)

type logSink struct {
	name      string
	keyValues map[string]interface{}
	logger    *log.Logger
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
// 0 = Error
func (l logSink) Info(level int, msg string, kvs ...interface{}) {
	var logLevel log.Level
	switch level {
	case 4:
		logLevel = log.Info
	case 8:
		logLevel = log.Debug
	case 1:
		logLevel = log.Warn
	case 0:
		logLevel = log.Error
	default:
		logLevel = log.Trace
	}

	logMsg := fmt.Sprintf("%s: %s", l.name, msg)
	for k, v := range l.keyValues {
		logMsg += fmt.Sprintf("%s: %+v  ", k, v)
	}
	for i := 0; i < len(kvs); i += 2 {
		logMsg += fmt.Sprintf("%s: %+v  ", kvs[i], kvs[i+1])
	}

	l.logger.Logf(logLevel, logMsg)
}

func (l logSink) Error(err error, msg string, kvs ...interface{}) {
	l.logger.Errorf(err, msg, kvs...)
}

func (l logSink) WithName(name string) logr.LogSink {
	return &logSink{
		name:      l.name + "." + name,
		keyValues: l.keyValues,
		logger:    l.logger,
	}
}

func (l logSink) WithValues(kvs ...interface{}) logr.LogSink {
	newMap := make(map[string]interface{}, len(l.keyValues)+len(kvs)/2)
	for k, v := range l.keyValues {
		newMap[k] = v
	}
	for i := 0; i < len(kvs); i += 2 {
		newMap[kvs[i].(string)] = kvs[i+1] //nolint:forcetypeassert
	}
	return &logSink{
		name:      l.name,
		keyValues: newMap,
		logger:    l.logger,
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
