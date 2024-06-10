package lsp

import (
	"os"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

var (
	snippetKind      = protocol.CompletionItemKindSnippet
	insertTextFormat = protocol.InsertTextFormatSnippet

	verbCompletionItem = protocol.CompletionItem{
		Label:  "ftl:verb",
		Kind:   &snippetKind,
		Detail: stringPtr("FTL Verb"),
		InsertText: stringPtr(`type ${1:Request} struct {}
type ${2:Response} struct{}

//ftl:verb
func ${3:Name}(ctx context.Context, req ${1:Request}) (${2:Response}, error) {
	return ${2:Response}{}, nil
}`),
		Documentation: &protocol.MarkupContent{
			Kind: protocol.MarkupKindMarkdown,
			Value: `Snippet for defining a verb function.
			
` + "```go" + `
//ftl:verb
func Name(ctx context.Context, req Request) (Response, error) {}
` + "```",
		},
		InsertTextFormat: &insertTextFormat,
	}

	enumTypeCompletionItem = protocol.CompletionItem{
		Label:  "ftl:enum (sum type)",
		Kind:   &snippetKind,
		Detail: stringPtr("FTL Enum (sum type)"),
		InsertText: stringPtr(`//ftl:enum
type ${1:Enum} string

const (
	${2:Value1} ${1:Enum} = "${2:Value1}"
	${3:Value2} ${1:Enum} = "${3:Value2}"
)`),
		Documentation: &protocol.MarkupContent{
			Kind: protocol.MarkupKindMarkdown,
			Value: `Snippet for defining a type enum.
		
` + "```go" + `
//ftl:enum
type MyEnum string

const (
	Value1 MyEnum = "Value1"
	Value2 MyEnum = "Value2"
)
` + "```",
		},
		InsertTextFormat: &insertTextFormat,
	}

	enumValueCompletionItem = protocol.CompletionItem{
		Label:  "ftl:enum (value)",
		Kind:   &snippetKind,
		Detail: stringPtr("FTL enum (value type)"),
		InsertText: stringPtr(`//ftl:enum
type ${1:Type} interface { ${2:interface}() }

type ${3:Value} struct {}
func (${3:Value}) ${2:interface}() {}
`),
		Documentation: &protocol.MarkupContent{
			Kind: protocol.MarkupKindMarkdown,
			Value: `Snippet for defining a value enum value.

` + "```go" + `
//ftl:enum
type Animal interface { animal() }

type Cat struct {}
func (Cat) animal() {}
` + "```",
		},
		InsertTextFormat: &insertTextFormat,
	}
)

var completionItems = []protocol.CompletionItem{
	verbCompletionItem,
	enumTypeCompletionItem,
	enumValueCompletionItem,
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
