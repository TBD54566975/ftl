package lsp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/puzpuzpuz/xsync/v3"
	_ "github.com/tliron/commonlog/simple"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	glspServer "github.com/tliron/glsp/server"
	"github.com/tliron/kutil/version"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/buildengine"
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
	documents   *documentStore
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
		documents:   newDocumentStore(),
	}

	handler.TextDocumentDidOpen = server.textDocumentDidOpen()
	handler.TextDocumentDidChange = server.textDocumentDidChange()
	handler.TextDocumentDidClose = server.textDocumentDidClose()
	handler.TextDocumentDidSave = server.textDocumentDidSave()
	handler.TextDocumentCompletion = server.textDocumentCompletion()
	handler.CompletionItemResolve = server.completionItemResolve()
	handler.TextDocumentHover = server.textDocumentHover()
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

type errSet []*schema.Error

// OnBuildStarted clears diagnostics for the given directory. New errors will arrive later if they still exist.
// Also emit an FTL message to set the status.
func (s *Server) OnBuildStarted(module buildengine.Module) {
	dirURI := "file://" + module.Config.Dir

	s.diagnostics.Range(func(uri protocol.DocumentUri, diagnostics []protocol.Diagnostic) bool {
		if strings.HasPrefix(uri, dirURI) {
			s.diagnostics.Delete(uri)
			s.publishDiagnostics(uri, []protocol.Diagnostic{})
		}
		return true
	})

	s.publishBuildState(buildStateBuilding, nil)
}

func (s *Server) OnBuildSuccess() {
	s.publishBuildState(buildStateSuccess, nil)
}

func (s *Server) OnBuildFailed(err error) {
	s.publishBuildState(buildStateFailure, err)
}

// Post sends diagnostics to the client.
func (s *Server) post(err error) {
	errByFilename := make(map[string]errSet)
	errUnspecified := []error{}

	// Deduplicate and associate by filename.
	for _, e := range ftlErrors.DeduplicateErrors(ftlErrors.UnwrapAll(err)) {
		if !ftlErrors.Innermost(e) {
			continue
		}
		var ce *schema.Error
		if errors.As(e, &ce) {
			filename := ce.Pos.Filename
			if _, exists := errByFilename[filename]; !exists {
				errByFilename[filename] = errSet{}
			}
			errByFilename[filename] = append(errByFilename[filename], ce)
		} else {
			errUnspecified = append(errUnspecified, err)
		}
	}

	go publishPositionalErrors(errByFilename, s)
	go publishUnspecifiedErrors(errUnspecified, s)
}

func publishPositionalErrors(errByFilename map[string]errSet, s *Server) {
	for filename, errs := range errByFilename {
		var diagnostics []protocol.Diagnostic
		for _, e := range errs {
			pp := e.Pos
			sourceName := "ftl"
			var severity protocol.DiagnosticSeverity

			switch e.Level {
			case schema.ERROR:
				severity = protocol.DiagnosticSeverityError
			case schema.WARN:
				severity = protocol.DiagnosticSeverityWarning
			case schema.INFO:
				severity = protocol.DiagnosticSeverityInformation
			}

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

// publishUnspecifiedErrors sends non-positional errors to the client as alerts.
func publishUnspecifiedErrors(errUnspecified []error, s *Server) {
	if s.glspContext == nil {
		return
	}

	for _, err := range errUnspecified {
		message := fmt.Sprintf("FTL Error: %s", err)

		go s.glspContext.Notify(protocol.ServerWindowShowMessage, protocol.ShowMessageParams{
			Type:    protocol.MessageTypeError,
			Message: message,
		})
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

type buildState string

const (
	buildStateBuilding buildState = "building"
	buildStateSuccess  buildState = "success"
	buildStateFailure  buildState = "failure"
)

type buildStateMessage struct {
	State buildState `json:"state"`
	Err   string     `json:"error,omitempty"`
}

func (s *Server) publishBuildState(state buildState, err error) {
	msg := buildStateMessage{State: state}
	if err != nil {
		msg.Err = err.Error()
	}

	s.logger.Debugf("Publishing build state: %s\n", msg)
	if s.glspContext == nil {
		return
	}

	go s.glspContext.Notify("ftl/buildState", msg)
}

func (s *Server) initialize() protocol.InitializeFunc {
	return func(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
		s.glspContext = context

		if params.Trace != nil {
			protocol.SetTraceValue(*params.Trace)
		}

		serverCapabilities := s.handler.CreateServerCapabilities()
		serverCapabilities.TextDocumentSync = protocol.TextDocumentSyncKindIncremental
		serverCapabilities.HoverProvider = true

		trueValue := true
		serverCapabilities.CompletionProvider = &protocol.CompletionOptions{
			ResolveProvider:   &trueValue,
			TriggerCharacters: []string{"/", "f"},
		}

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

func (s *Server) textDocumentDidOpen() protocol.TextDocumentDidOpenFunc {
	return func(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
		uri := params.TextDocument.URI
		content := params.TextDocument.Text
		s.documents.set(uri, content)
		return nil
	}
}

func (s *Server) textDocumentDidChange() protocol.TextDocumentDidChangeFunc {
	return func(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
		doc, ok := s.documents.get(params.TextDocument.URI)
		if !ok {
			return nil
		}

		doc.update(params.ContentChanges)
		return nil
	}
}

func (s *Server) textDocumentDidClose() protocol.TextDocumentDidCloseFunc {
	return func(context *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
		uri := params.TextDocument.URI
		s.documents.delete(uri)
		return nil
	}
}

func (s *Server) textDocumentDidSave() protocol.TextDocumentDidSaveFunc {
	return func(context *glsp.Context, params *protocol.DidSaveTextDocumentParams) error {
		return nil
	}
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

			// Define a custom function to check for spaces or special characters
			isDelimiter := func(char rune) bool {
				switch char {
				case ' ', '\t', '[', ']', '{', '}', '(', ')':
					return true
				default:
					return false
				}
			}

			end := start
			for end < len(lineText) && !isDelimiter(rune(lineText[end])) {
				end++
			}

			// If starting column is out of range, return 0
			if start >= len(lineText) {
				return 0, nil
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
