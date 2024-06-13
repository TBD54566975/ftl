package lsp

import (
	_ "embed"
	"os"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

//go:embed markdown/completion/verb.md
var verbCompletionDocs string

//go:embed markdown/completion/enumType.md
var enumTypeCompletionDocs string

//go:embed markdown/completion/enumValue.md
var enumValueCompletionDocs string

var completionItems = []protocol.CompletionItem{
	completionItem("ftl:verb", "FTL Verb", verbCompletionDocs),
	completionItem("ftl:enum (sum type)", "FTL Enum (sum type)", enumTypeCompletionDocs),
	completionItem("ftl:enum (value)", "FTL Enum (value type)", enumValueCompletionDocs),
}

func completionItem(label, detail, markdown string) protocol.CompletionItem {
	snippetKind := protocol.CompletionItemKindSnippet
	insertTextFormat := protocol.InsertTextFormatSnippet

	// Split markdown by "---"
	// First half is completion docs, second half is insert text
	parts := strings.Split(markdown, "---")
	if len(parts) != 2 {
		panic("invalid markdown. must contain exactly one '---' to separate completion docs from insert text")
	}

	insertText := strings.TrimSpace(parts[1])
	return protocol.CompletionItem{
		Label:      label,
		Kind:       &snippetKind,
		Detail:     &detail,
		InsertText: &insertText,
		Documentation: &protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: strings.TrimSpace(parts[0]),
		},
		InsertTextFormat: &insertTextFormat,
	}
}

func (s *Server) textDocumentCompletion() protocol.TextDocumentCompletionFunc {
	return func(context *glsp.Context, params *protocol.CompletionParams) (interface{}, error) {
		uri := params.TextDocument.URI
		position := params.Position

		doc, ok := s.documents.get(uri)
		if !ok {
			return nil, nil
		}

		line := int(position.Line - 1)
		if line >= len(doc.lines) {
			return nil, nil
		}

		lineContent := doc.lines[line]
		character := int(position.Character - 1)
		if character > len(lineContent) {
			character = len(lineContent)
		}

		prefix := lineContent[:character]

		// Filter completion items based on the prefix
		var filteredItems []protocol.CompletionItem
		for _, item := range completionItems {
			if strings.HasPrefix(item.Label, prefix) || strings.Contains(item.Label, prefix) {
				filteredItems = append(filteredItems, item)
			}
		}

		return &protocol.CompletionList{
			IsIncomplete: false,
			Items:        filteredItems,
		}, nil
	}
}

func (s *Server) completionItemResolve() protocol.CompletionItemResolveFunc {
	return func(context *glsp.Context, params *protocol.CompletionItem) (*protocol.CompletionItem, error) {
		if path, ok := params.Data.(string); ok {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil, err
			}

			params.Documentation = &protocol.MarkupContent{
				Kind:  protocol.MarkupKindMarkdown,
				Value: string(content),
			}
		}

		return params, nil
	}
}

func stringPtr(v string) *string {
	s := v
	return &s
}
