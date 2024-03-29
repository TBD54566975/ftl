package compile

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"github.com/TBD54566975/ftl/backend/schema"
)

// This file contains a parser for Go FTL directives.
//
// eg. //ftl:ingress GET /foo/bar

type directiveWrapper struct {
	Directive directive `parser:"'ftl' ':' @@"`
}

//sumtype:decl
type directive interface{ directive() }

type directiveExport struct {
	Pos lexer.Position

	Export bool `parser:"@'export'"`
}

func (*directiveExport) directive()       {}
func (d *directiveExport) String() string { return "ftl:export" }

type directiveIngress struct {
	Pos schema.Position

	Type   string                        `parser:"'ingress' @('http')?"`
	Method string                        `parser:"@('GET' | 'POST' | 'PUT' | 'DELETE')"`
	Path   []schema.IngressPathComponent `parser:"('/' @@)+"`
}

func (*directiveIngress) directive() {}
func (d *directiveIngress) String() string {
	w := &strings.Builder{}
	fmt.Fprintf(w, "ftl:ingress %s", d.Method)
	for _, p := range d.Path {
		fmt.Fprintf(w, "/%s", p)
	}
	return w.String()
}

var directiveParser = participle.MustBuild[directiveWrapper](
	participle.Lexer(schema.Lexer),
	participle.Elide("Whitespace"),
	participle.Unquote(),
	participle.UseLookahead(2),
	participle.Union[directive](&directiveExport{}, &directiveIngress{}),
	participle.Union[schema.IngressPathComponent](&schema.IngressPathLiteral{}, &schema.IngressPathParameter{}),
)

func parseDirectives(fset *token.FileSet, docs *ast.CommentGroup) ([]directive, error) {
	if docs == nil {
		return nil, nil
	}
	directives := []directive{}
	for _, line := range docs.List {
		if !strings.HasPrefix(line.Text, "//ftl:") {
			continue
		}
		pos := fset.Position(line.Pos())
		// TODO: We need to adjust position information embedded in the schema.
		directive, err := directiveParser.ParseString(pos.Filename, line.Text[2:])
		if err != nil {
			// Adjust the Participle-reported position relative to the AST node.
			var perr participle.Error
			if errors.As(err, &perr) {
				ppos := perr.Position()
				ppos.Filename = pos.Filename
				ppos.Column += pos.Column + 2
				ppos.Line = pos.Line
				err = participle.Errorf(ppos, "%s", perr.Message())
			} else {
				err = fmt.Errorf("%s: %w", pos, err)
			}
			return nil, fmt.Errorf("%s: %w", "invalid directive", err)
		}
		directives = append(directives, directive.Directive)
	}
	return directives, nil
}
