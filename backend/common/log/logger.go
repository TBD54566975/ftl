package log

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"golang.org/x/exp/maps"
)

var _ Interface = (*Logger)(nil)

const scopeKey = "scope"

type Entry struct {
	Time       time.Time         `json:"-"`
	Level      Level             `json:"level"`
	Attributes map[string]string `json:"attributes,omitempty"`
	Message    string            `json:"message"`

	Error error `json:"-"`
}

// Logger is the concrete logger.
type Logger struct {
	level      Level
	attributes map[string]string
	sink       Sink
}

// New returns a new logger.
func New(level Level, sink Sink) *Logger {
	return &Logger{
		level:      level,
		attributes: map[string]string{},
		sink:       sink,
	}
}

func (l Logger) Scope(scope string) *Logger {
	return l.Sub(map[string]string{scopeKey: scope})
}

// Sub creates a new logger with the given attributes.
func (l Logger) Sub(attributes map[string]string) *Logger {
	attr := map[string]string{}
	maps.Copy(attr, l.attributes)
	maps.Copy(attr, attributes)
	l.attributes = attr
	return &l
}

func (l Logger) AddSink(sink Sink) *Logger {
	l.sink = Tee(l.sink, sink)
	return &l
}

func (l Logger) Level(level Level) *Logger {
	l.level = level
	return &l
}

func (l Logger) GetLevel() Level {
	return l.level
}

func (l *Logger) Log(entry Entry) {
	if entry.Level < l.level {
		return
	}
	if entry.Time.IsZero() {
		entry.Time = time.Now()
	}
	entry.Attributes = l.attributes
	if err := l.sink.Log(entry); err != nil {
		fmt.Fprintf(os.Stderr, "ftl:log: failed to log entry: %v", err)
	}
}

func (l *Logger) Logf(level Level, format string, args ...interface{}) {
	l.Log(Entry{Level: level, Message: fmt.Sprintf(format, args...)})
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
	if err == nil {
		return
	}
	l.Log(Entry{Level: Error, Message: fmt.Sprintf(format, args...) + ": " + err.Error(), Error: err})
}

// WriterAt returns a writer that logs each line at the given level.
func (l *Logger) WriterAt(level Level) *io.PipeWriter {
	// Based on MIT licensed Logrus https://github.com/sirupsen/logrus/blob/bdc0db8ead3853c56b7cd1ac2ba4e11b47d7da6b/writer.go#L27
	reader, writer := io.Pipe()
	var printFunc func(format string, args ...interface{})

	switch level {
	case Trace:
		printFunc = l.Tracef
	case Debug:
		printFunc = l.Debugf
	case Info:
		printFunc = l.Infof
	case Warn:
		printFunc = l.Warnf
	case Error:
		printFunc = func(format string, args ...interface{}) {
			l.Errorf(nil, format, args...)
		}
	default:
		panic(level)
	}

	go l.writerScanner(reader, printFunc)
	runtime.SetFinalizer(writer, writerFinalizer)

	return writer
}

func (l *Logger) writerScanner(reader *io.PipeReader, printFunc func(format string, args ...interface{})) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(nil, 1024*1024) // 1MB buffer
	for scanner.Scan() {
		printFunc("%s", scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		l.Errorf(err, "Error while reading from Writer")
	}
	reader.Close()
}

func writerFinalizer(writer *io.PipeWriter) {
	writer.Close()
}
