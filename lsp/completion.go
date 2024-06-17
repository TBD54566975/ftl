package lsp

import (
	_ "embed"
	"fmt"
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

//go:embed markdown/completion/typeAlias.md
var typeAliasCompletionDocs string

//go:embed markdown/completion/ingress.md
var ingressCompletionDocs string

//go:embed markdown/completion/cron.md
var cronCompletionDocs string

//go:embed markdown/completion/cronExpression.md
var cronExpressionCompletionDocs string

//go:embed markdown/completion/retry.md
var retryCompletionDocs string

//go:embed markdown/completion/configDeclare.md
var declareConfigCompletionDocs string

//go:embed markdown/completion/secretDeclare.md
var declareSecretCompletionDocs string

//go:embed markdown/completion/pubSubTopic.md
var declarePubSubTopicCompletionDocs string

//go:embed markdown/completion/pubSubSubscription.md
var declarePubSubSubscriptionCompletionDocs string

//go:embed markdown/completion/pubSubSink.md
var definePubSubSinkCompletionDocs string

//go:embed markdown/completion/fsmDeclare.md
var fsmCompletionDocs string

// Markdown is split by "---". First half is completion docs, second half is insert text.
var completionItems = []protocol.CompletionItem{
	completionItem("ftl:verb", "FTL Verb", verbCompletionDocs),
	completionItem("ftl:enum (sum type)", "FTL Enum (sum type)", enumTypeCompletionDocs),
	completionItem("ftl:enum (value)", "FTL Enum (value type)", enumValueCompletionDocs),
	completionItem("ftl:typealias", "FTL Type Alias", typeAliasCompletionDocs),
	completionItem("ftl:ingress", "FTL Ingress", ingressCompletionDocs),
	completionItem("ftl:cron", "FTL Cron", cronCompletionDocs),
	completionItem("ftl:cron:expression", "FTL Cron with expression", cronExpressionCompletionDocs),
	completionItem("ftl:retry", "FTL Retry", retryCompletionDocs),
	completionItem("ftl:config:declare", "Declare config", declareConfigCompletionDocs),
	completionItem("ftl:secret:declare", "Declare secret", declareSecretCompletionDocs),
	completionItem("ftl:pubsub:topic", "Declare PubSub topic", declarePubSubTopicCompletionDocs),
	completionItem("ftl:pubsub:subscription", "Declare a PubSub subscription", declarePubSubSubscriptionCompletionDocs),
	completionItem("ftl:pubsub:sink", "Define a PubSub sink", definePubSubSinkCompletionDocs),
	completionItem("ftl:fsm", "Model a FSM", fsmCompletionDocs),
}

// Track which directives are //ftl: prefixed, so the we can autocomplete them via `/`.
// This is built at init time and does not change during runtime.
var directiveItems = map[string]bool{}

func completionItem(label, detail, markdown string) protocol.CompletionItem {
	snippetKind := protocol.CompletionItemKindSnippet
	insertTextFormat := protocol.InsertTextFormatSnippet

	parts := strings.Split(markdown, "---")
	if len(parts) != 2 {
		panic(fmt.Sprintf("completion item %q: invalid markdown. must contain exactly one '---' to separate completion docs from insert text", label))
	}

	insertText := strings.TrimSpace(parts[1])
	// Warn if we see two spaces in the insert text.
	if strings.Contains(insertText, "  ") {
		panic(fmt.Sprintf("completion item %q: contains two spaces in the insert text. Use tabs instead!", label))
	}

	// If there is a `//ftl:` this can be autocompleted when the user types `/`.
	if strings.Contains(insertText, "//ftl:") {
		directiveItems[label] = true
	}

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

		// Line and Character are 0-based, however the cursor can be after the last character in the line.
		line := int(position.Line)
		if line >= len(doc.lines) {
			return nil, nil
		}
		lineContent := doc.lines[line]
		character := int(position.Character)
		if character > len(lineContent) {
			character = len(lineContent)
		}

		// Currently all completions are in global scope, so the completion must be triggered at the beginning of the line.
		// To do this, check to the start of the line and if there is any whitespace, it is not completing a whole word from the start.
		// We also want to check that the cursor is at the end of the line so we dont let stray chars shoved at the end of the completion.
		isAtEOL := character == len(lineContent)
		if !isAtEOL {
			return nil, nil
		}

		// Is not completing from the start of the line.
		if strings.ContainsAny(lineContent, " \t") {
			return nil, nil
		}

		// If there is a single `/` at the start of the line, we can autocomplete directives. eg `/f`.
		// This is a hint to the user that these are ftl directives.
		// Note that what I can tell, VSCode won't trigger completion on and after `//` so we can only complete on half of a comment.
		isSlashed := strings.HasPrefix(lineContent, "/")
		if isSlashed {
			lineContent = strings.TrimPrefix(lineContent, "/")
		}

		// Filter completion items based on the line content and if it is a directive.
		var filteredItems []protocol.CompletionItem
		for _, item := range completionItems {
			if !strings.Contains(item.Label, lineContent) {
				continue
			}

			if isSlashed && !directiveItems[item.Label] {
				continue
			}

			if isSlashed {
				// Remove that / from the start of the line, so that the completion doesn't have `///`.
				// VSCode doesn't seem to want to remove the `/` for us.
				item.AdditionalTextEdits = []protocol.TextEdit{
					{
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      uint32(line),
								Character: 0,
							},
							End: protocol.Position{
								Line:      uint32(line),
								Character: 1,
							},
						},
						NewText: "",
					},
				}
			}

			filteredItems = append(filteredItems, item)
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
