package observability

import (
	"fmt"

	"github.com/go-logr/logr"

	"github.com/TBD54566975/ftl/internal/log"
)

type logSink struct {
	keyValues map[string]interface{}
	logger    *log.Logger
}

func NewOtelLogger(logger *log.Logger, level log.Level) logr.Logger {
	sink := &logSink{
		logger: logger.Scope("otel").Level(level),
	}
	return logr.New(sink)
}

var _ logr.LogSink = &logSink{}

func (l *logSink) Init(info logr.RuntimeInfo) {
}

func (l logSink) Enabled(level int) bool {
	return otelLevelToLevel(level) >= l.logger.GetLevel()
}

func (l logSink) Info(level int, msg string, kvs ...interface{}) {
	// otel uses the following mapping for level to our log.Level
	// 8 = Debug
	// 4 = Info
	// 1 = Warning
	// 0 = Error
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

	logMsg := msg + " "
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
		keyValues: l.keyValues,
		logger:    l.logger.Scope(name),
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
		keyValues: newMap,
		logger:    l.logger,
	}
}

func otelLevelToLevel(level int) log.Level {
	switch level {
	case 4:
		return log.Info
	case 8:
		return log.Debug
	case 1:
		return log.Warn
	case 0:
		return log.Error
	default:
		return log.Trace
	}
}
