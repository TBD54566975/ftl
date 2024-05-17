package log

import (
	"io"
)

// Config for the logger.
type Config struct {
	Level      Level `help:"Log level." default:"info" env:"LOG_LEVEL"`
	JSON       bool  `help:"Log in JSON format." env:"LOG_JSON"`
	Timestamps bool  `help:"Include timestamps in text logs." env:"LOG_TIMESTAMPS"`
	Color      bool  `help:"Enable colored output regardless of TTY." env:"LOG_COLOR"`
}

// Configure returns a new logger based on the config.
func Configure(w io.Writer, cfg Config) *Logger {
	var sink Sink
	if cfg.JSON {
		sink = newJSONSink(w)
	} else {
		sink = newPlainSink(w, cfg.Timestamps, cfg.Color)
	}
	return New(cfg.Level, sink)
}
