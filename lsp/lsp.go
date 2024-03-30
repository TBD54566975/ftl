package lsp

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	_ "github.com/tliron/commonlog/simple"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	glspServer "github.com/tliron/glsp/server"
	"github.com/tliron/kutil/version"

	"github.com/TBD54566975/ftl/backend/schema"
	ftlErrors "github.com/TBD54566975/ftl/internal/errors"
	"github.com/TBD54566975/ftl/internal/log"
)

const lsName = "ftl-language-server"

// Server is a language server.
type Server struct {
	server      *glspServer.Server
	glspLogger  *GLSPLogger
	glspContext *glsp.Context
	handler     protocol.Handler
	logger      log.Logger
}

// NewServer creates a new language server.
func NewServer(ctx context.Context) *Server {
	handler := protocol.Handler{
		Initialized: initialized,
		Shutdown:    shutdown,
		SetTrace:    setTrace,
		LogTrace:    logTrace,
	}
	s := glspServer.NewServer(&handler, lsName, false)
	server := &Server{
		server:     s,
		glspLogger: NewGLSPLogger(s.Log),
		logger:     *log.FromContext(ctx),
	}
	handler.Initialize = server.initialize()
	// handler.TextDocumentDidOpen = server.textDocumentDidOpen()
	handler.TextDocumentDidChange = server.textDocumentDidChange()
	return server
}

func (s *Server) Run() error {
	return errors.Wrap(s.server.RunStdio(), "lsp")
}

type errSet map[string]schema.Error

// Post sends diagnostics to the client. err must be joined schema.Errors.
func (s *Server) Post(err error) {
	errByFilename := make(map[string]errSet)

	// Deduplicate and associate by filename.
	for _, subErr := range ftlErrors.UnwrapAll(err) {
		var ce schema.Error
		if errors.As(subErr, &ce) {
			filename := ce.Pos.Filename
			if _, exists := errByFilename[filename]; !exists {
				errByFilename[filename] = make(errSet)
			}
			errByFilename[filename][strings.TrimSpace(ce.Error())] = ce
		}
	}

	go func() {
		for filename, errs := range errByFilename {
			var diagnostics []protocol.Diagnostic
			for _, e := range errs {
				pp := e.Pos
				sourceName := "ftl"
				severity := protocol.DiagnosticSeverityError
				diagnostics = append(diagnostics, protocol.Diagnostic{
					Range: protocol.Range{
						Start: protocol.Position{Line: uint32(pp.Line - 1), Character: uint32(pp.Column - 1)},
						End:   protocol.Position{Line: uint32(pp.Line - 1), Character: uint32(pp.Column + 10 - 1)}, //todo: fix
					},
					Severity: &severity,
					Source:   &sourceName,
					Message:  e.Msg,
				})
			}

			if s.glspContext == nil {
				return
			}

			go s.glspContext.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
				URI:         "file://" + filename,
				Diagnostics: diagnostics,
			})
		}
	}()
}

func (s *Server) initialize() protocol.InitializeFunc {
	return func(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
		s.glspContext = context

		if params.Trace != nil {
			protocol.SetTraceValue(*params.Trace)
		}

		serverCapabilities := s.handler.CreateServerCapabilities()
		return protocol.InitializeResult{
			Capabilities: serverCapabilities,
			ServerInfo: &protocol.InitializeResultServerInfo{
				Name:    lsName,
				Version: &version.GitVersion,
			},
		}, nil
	}
}

func (s *Server) textDocumentDidChange() protocol.TextDocumentDidChangeFunc {
	return func(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
		// s.refreshDiagnosticsOfDocument(params.TextDocument.URI)
		return nil
	}
}

func (s *Server) clearDiagnosticsOfDocument(uri protocol.DocumentUri) {
	go func() {
		go s.glspContext.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
			URI:         uri,
			Diagnostics: []protocol.Diagnostic{},
		})
	}()
}

func initialized(context *glsp.Context, params *protocol.InitializedParams) error {
	return nil
}

func shutdown(context *glsp.Context) error {
	protocol.SetTraceValue(protocol.TraceValueOff)
	return nil
}

func logTrace(context *glsp.Context, params *protocol.LogTraceParams) error {
	return nil
}

func setTrace(context *glsp.Context, params *protocol.SetTraceParams) error {
	protocol.SetTraceValue(params.Value)
	return nil
}
