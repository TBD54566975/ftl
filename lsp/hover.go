package lsp

import (
	_ "embed"
	"strings"

	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
)

//go:embed markdown/hover/verb.md
var verbHoverContent string

//go:embed markdown/hover/enum.md
var enumHoverContent string

var hoverMap = map[string]string{
	"//ftl:verb": verbHoverContent,
	"//ftl:enum": enumHoverContent,
}

func (s *Server) textDocumentHover() protocol.TextDocumentHoverFunc {
	return func(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
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
		character := int(position.Character)
		if character > len(lineContent) {
			character = len(lineContent)
		}

		for hoverString, hoverContent := range hoverMap {
			startIndex := strings.Index(lineContent, hoverString)
			if startIndex != -1 && startIndex <= character && character <= startIndex+len(hoverString) {
				return &protocol.Hover{
					Contents: &protocol.MarkupContent{
						Kind:  protocol.MarkupKindMarkdown,
						Value: hoverContent,
					},
				}, nil
			}
		}

		return nil, nil
	}
}
