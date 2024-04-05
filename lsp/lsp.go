package lsp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/puzpuzpuz/xsync/v3"
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
	glspContext *glsp.Context
	handler     protocol.Handler
	logger      log.Logger
	diagnostics *xsync.MapOf[protocol.DocumentUri, []protocol.Diagnostic]
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
		server:      s,
		logger:      *log.FromContext(ctx).Scope("lsp"),
		diagnostics: xsync.NewMapOf[protocol.DocumentUri, []protocol.Diagnostic](),
	}
	handler.Initialize = server.initialize()
	return server
}

func (s *Server) Run() error {
	err := s.server.RunStdio()
	if err != nil {
		return fmt.Errorf("lsp: %w", err)
	}
	return nil
}

type errSet map[string]schema.Error

// BuildStarted clears diagnostics for the given directory. New errors will arrive later if they still exist.
func (s *Server) BuildStarted(dir string) {
	dirURI := "file://" + dir

	s.diagnostics.Range(func(uri protocol.DocumentUri, diagnostics []protocol.Diagnostic) bool {
		if strings.HasPrefix(uri, dirURI) {
			s.diagnostics.Delete(uri)
			s.publishDiagnostics(uri, []protocol.Diagnostic{})
		}
		return true
	})
}

// Post sends diagnostics to the client. err must be joined schema.Errors.
func (s *Server) post(err error) {
	errByFilename := make(map[string]errSet)

	// Deduplicate and associate by filename.
	errs := ftlErrors.UnwrapAll(err)
	for _, err := range errs {
		if ftlErrors.Innermost(err) {
			var ce schema.Error
			if errors.As(err, &ce) {
				filename := ce.Pos.Filename
				if _, exists := errByFilename[filename]; !exists {
					errByFilename[filename] = make(errSet)
				}
				errByFilename[filename][strings.TrimSpace(ce.Error())] = ce
			}
		}
	}

	go publishErrors(errByFilename, s)
}

func publishErrors(errByFilename map[string]errSet, s *Server) {
	for filename, errs := range errByFilename {
		var diagnostics []protocol.Diagnostic
		for _, e := range errs {
			pp := e.Pos
			sourceName := "ftl"
			severity := protocol.DiagnosticSeverityError

			// If the end column is not set, set it to the length of the word.
			if e.EndColumn <= pp.Column {
				length, err := getLineOrWordLength(filename, pp.Line, pp.Column, false)
				if err != nil {
					s.logger.Errorf(err, "Failed to get line or word length")
					continue
				}
				e.EndColumn = pp.Column + length
			}

			diagnostics = append(diagnostics, protocol.Diagnostic{
				Range: protocol.Range{
					Start: protocol.Position{Line: uint32(pp.Line - 1), Character: uint32(pp.Column - 1)},
					End:   protocol.Position{Line: uint32(pp.Line - 1), Character: uint32(e.EndColumn - 1)},
				},
				Severity: &severity,
				Source:   &sourceName,
				Message:  e.Msg,
			})
		}

		uri := "file://" + filename
		s.diagnostics.Store(uri, diagnostics)
		s.publishDiagnostics(uri, diagnostics)
	}
}

func (s *Server) publishDiagnostics(uri protocol.DocumentUri, diagnostics []protocol.Diagnostic) {
	s.logger.Debugf("Publishing diagnostics for %s\n", uri)
	if s.glspContext == nil {
		return
	}

	go s.glspContext.Notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
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

// getLineOrWordLength returns the length of the line or the length of the word starting at the given column.
// If wholeLine is true, it returns the length of the entire line.
// If wholeLine is false, it returns the length of the word starting at the column.
func getLineOrWordLength(filePath string, lineNum, column int, wholeLine bool) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentLine := 1
	for scanner.Scan() {
		if currentLine == lineNum {
			lineText := scanner.Text()
			if wholeLine {
				return len(lineText), nil
			}
			start := column - 1
			end := start
			for end < len(lineText) && !unicode.IsSpace(rune(lineText[end])) {
				end++
			}
			return end - start, nil
		}
		currentLine++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return 0, os.ErrNotExist
}
