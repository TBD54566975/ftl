package log

import (
	"io"
)

// Config for the logger.
type Config struct {
	Level Level `help:"Log level." default:"info" env:"LOG_LEVEL"`
	JSON  bool  `help:"Log in JSON format."`
}

// Configure returns a new logger based on the config.
func Configure(w io.Writer, cfg Config) *Logger {
	var sink Sink
	if cfg.JSON {
		sink = newJSONSink(w)
	} else {
		sink = newPlainSink(w)
	}
	return New(cfg.Level, sink)
}
