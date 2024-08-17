package lsp

import (
	"github.com/TBD54566975/ftl/internal/log"
)

type LogSink struct {
	server *Server
}

var _ log.Sink = (*LogSink)(nil)

func NewLogSink(server *Server) *LogSink {
	return &LogSink{server: server}
}

func (l *LogSink) Log(entry log.Entry) error {
	switch entry.Level {
	case log.Error:
		l.server.post(entry.Error)
	case log.Warn:
		if entry.Error != nil {
			l.server.post(entry.Error)
		}
	default:
	}
	return nil
}
