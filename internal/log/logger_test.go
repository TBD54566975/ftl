package log

import (
	"errors"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/benbjohnson/clock"
)

func TestLogger(t *testing.T) {
	w := &strings.Builder{}
	log := New(Trace, newJSONSink(w))
	log.clock = clock.NewMock()
	log.Tracef("trace: %s", "trace")
	log.Debugf("debug: %s", "debug")
	log.Infof("info: %s", "info")
	log.Warnf("warn: %s", "warn")
	log.Errorf(errors.New("error"), "error: %s", "error")
	log = log.Scope("scoped").Attrs(map[string]string{"key": "value"})
	log.Tracef("trace: %s", "trace")
	log.Log(Entry{Level: Trace, Message: "trace: trace"})
	assert.Equal(t, strings.TrimSpace(`
{"level":"trace","message":"trace: trace","time":"1970-01-01T00:00:00Z"}
{"level":"debug","message":"debug: debug","time":"1970-01-01T00:00:00Z"}
{"level":"info","message":"info: info","time":"1970-01-01T00:00:00Z"}
{"level":"warn","message":"warn: warn","time":"1970-01-01T00:00:00Z"}
{"level":"error","message":"error: error: error","time":"1970-01-01T00:00:00Z","error":"error"}
{"level":"trace","attributes":{"key":"value","scope":"scoped"},"message":"trace: trace","time":"1970-01-01T00:00:00Z"}
{"level":"trace","attributes":{"key":"value","scope":"scoped"},"message":"trace: trace","time":"1970-01-01T00:00:00Z"}
`)+"\n", w.String())
}
