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
	completionItem("ftl:retry", "FTL Retry", retryCompletionDocs),
	completionItem("ftl:config:declare", "Declare config", declareConfigCompletionDocs),
	completionItem("ftl:secret:declare", "Declare secret", declareSecretCompletionDocs),
	completionItem("ftl:pubsub:topic", "Declare PubSub topic", declarePubSubTopicCompletionDocs),
	completionItem("ftl:pubsub:subscription", "Declare a PubSub subscription", declarePubSubSubscriptionCompletionDocs),
	completionItem("ftl:pubsub:sink", "Define a PubSub sink", definePubSubSinkCompletionDocs),
	completionItem("ftl:fsm", "Model a FSM", fsmCompletionDocs),
}

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
		fmt.Fprintf(os.Stderr, "warning: completion item %q contains two spaces in the insert text. Use tabs instead!\n", label)
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

		line := int(position.Line)
		if line >= len(doc.lines) {
			return nil, nil
		}

		lineContent := doc.lines[line]
		character := int(position.Character - 1)
		if character > len(lineContent) {
			character = len(lineContent)
		}

		prefix := lineContent[character:]

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
