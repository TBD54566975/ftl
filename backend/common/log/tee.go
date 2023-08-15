package log

import (
	"github.com/alecthomas/errors"
)

// Tee returns a sink that writes to all of the given sinks.
func Tee(sinks ...Sink) Sink {
	return &tee{sinks: sinks}
}

type tee struct {
	sinks []Sink
}

func (t *tee) Log(entry Entry) error {
	var errs []error
	for _, sink := range t.sinks {
		if err := sink.Log(entry); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
