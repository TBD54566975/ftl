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

// Track which directives are //ftl: prefixed, so the we can autocomplete them via `//`.
var directiveItems = map[string]bool{}

func completionItem(label, detail, markdown string) protocol.CompletionItem {
	snippetKind := protocol.CompletionItemKindSnippet
	insertTextFormat := protocol.InsertTextFormatSnippet

	parts := strings.Split(markdown, "---")
	if len(parts) != 2 {
		panic("invalid markdown. must contain exactly one '---' to separate completion docs from insert text")
	}

	insertText := strings.TrimSpace(parts[1])
	// Warn if we see two spaces in the insert text.
	if strings.Contains(insertText, "  ") {
		panic(fmt.Sprintf("completion item %q contains two spaces in the insert text. Use tabs instead!", label))
	}

	// If there is a `//ftl:` this can be autocompleted when the user types `//`.
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
		fmt.Fprintf(os.Stderr, "textDocument/completion: %v\n", params)

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
		// To do this, check to the start of the line and if there is any whitespace, it is not completing from the start.
		// We also want to check that the cursor is at the end of the line so we dont stomp over existing text.
		isAtEOL := character == len(lineContent)
		if !isAtEOL {
			fmt.Fprintf(os.Stderr, "not at end of line character: %d lineContent: %q len(lineContent): %d\n", character, lineContent, len(lineContent))

			return &protocol.CompletionList{
				IsIncomplete: false,
				Items:        []protocol.CompletionItem{},
			}, nil
		}

		// Is not completing from the start of the line.
		if strings.ContainsAny(lineContent, " \t") {
			fmt.Fprintf(os.Stderr, "not at start of line\n")

			return &protocol.CompletionList{
				IsIncomplete: false,
				Items:        []protocol.CompletionItem{},
			}, nil
		}

		// If there is a single `/` at the start of the line, we can autocomplete directives. eg `/f`.
		// This is a hint to the user that these are ftl directives.
		// Note that what I can tell, VSCode won't trigger completion on and after `//` so we can only complete on half of a comment.
		isDirective := strings.HasPrefix(lineContent, "/")
		if isDirective {
			lineContent = strings.TrimPrefix(lineContent, "/")
		}

		// Filter completion items based on the line content and if it is a directive.
		var filteredItems []protocol.CompletionItem
		for _, item := range completionItems {
			if !strings.Contains(item.Label, lineContent) {
				fmt.Fprintf(os.Stderr, "skipping item %q\n", item.Label)
				continue
			}

			if isDirective && !directiveItems[item.Label] {
				fmt.Fprintf(os.Stderr, "skipping directive item %q\n", item.Label)
				continue
			}

			fmt.Fprintf(os.Stderr, "adding item %q\n", item.Label)
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
		fmt.Fprintf(os.Stderr, "completionItem/resolve: %v\n", params)

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
