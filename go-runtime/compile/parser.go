package compile

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/alecthomas/participle/v2"

	"github.com/TBD54566975/ftl/backend/schema"
)

// This file contains a parser for Go FTL directives.
//
// eg. //ftl:public http GET /foo/bar

type directiveWrapper struct {
	Directive directive `parser:"'ftl' ':' @@"`
}

//sumtype:decl
type directive interface{ directive() }

type directiveVisibility struct {
	Pos schema.Position `parser:"" protobuf:"1,optional"`

	Visibility schema.Visibility             `parser:"@('public' | 'internal' | 'private')"`
	Type       string                        `parser:"@('http')?"`
	Method     string                        `parser:"@('GET' | 'POST' | 'PUT' | 'DELETE')?"`
	Path       []schema.IngressPathComponent `parser:"('/' @@)*"`
}

func (*directiveVisibility) directive() {}
func (d *directiveVisibility) String() string {
	w := &strings.Builder{}
	fmt.Fprintf(w, "ftl:%s", d.Visibility)
	if d.Type != "" {
		fmt.Fprintf(w, " %s", d.Type)
	}
	if d.Method != "" {
		fmt.Fprintf(w, " %s", d.Method)
	}
	for _, p := range d.Path {
		fmt.Fprintf(w, "/%s", p)
	}
	return w.String()
}

type directiveCronJob struct {
	Pos schema.Position

	Cron string `parser:"'cron' Whitespace @((' ' | Number | '-' | '/' | '*' | ',')+)"`
}

func (*directiveCronJob) directive() {}

func (d *directiveCronJob) String() string {
	return fmt.Sprintf("cron %s", d.Cron)
}

var directiveParser = participle.MustBuild[directiveWrapper](
	participle.Lexer(schema.Lexer),
	participle.Elide("Whitespace"),
	participle.Unquote(),
	participle.UseLookahead(2),
	participle.Union[directive](&directiveVisibility{}, &directiveCronJob{}),
	participle.Union[schema.IngressPathComponent](&schema.IngressPathLiteral{}, &schema.IngressPathParameter{}),
)

func parseDirectives(node ast.Node, fset *token.FileSet, docs *ast.CommentGroup) ([]directive, *schema.Error) {
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
			var scerr *schema.Error
			var perr participle.Error
			if errors.As(err, &perr) {
				ppos := schema.Position{}
				ppos.Filename = pos.Filename
				ppos.Column += pos.Column + 2
				ppos.Line = pos.Line
				scerr = schema.Errorf(ppos, ppos.Column, "%s", perr.Message())
			} else {
				scerr = wrapf(node, err, "")
			}
			return nil, scerr
		}
		directives = append(directives, directive.Directive)
	}
	return directives, nil
}
