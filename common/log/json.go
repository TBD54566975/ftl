package log

import (
	"encoding/json"
	"io"

	"github.com/alecthomas/errors"
)

var _ Sink = (*jsonSink)(nil)

type jsonEntry struct {
	Entry
	Error string `json:"error,omitempty"`
}

func newJSONSink(w io.Writer) *jsonSink {
	return &jsonSink{
		w:   w,
		enc: json.NewEncoder(w),
	}
}

type jsonSink struct {
	w   io.Writer
	enc *json.Encoder
}

func (j *jsonSink) Log(entry Entry) error {
	defer j.w.Write([]byte{'\n'}) //nolint
	jentry := jsonEntry{
		Error: entry.Error.Error(),
		Entry: entry,
	}
	return errors.WithStack(j.enc.Encode(jentry))
}
