package compile

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// This file contains a parser for Go FTL directives.
//
// eg. //ftl:ingress GET /foo/bar

type directiveWrapper struct {
	Directive directive `parser:"'ftl' ':' @@"`
}

//sumtype:decl
type directive interface{ directive() }

type directiveVerb struct {
	Pos lexer.Position

	Verb bool `parser:"@'verb'"`
}

func (*directiveVerb) directive()       {}
func (d *directiveVerb) String() string { return "ftl:verb" }

type directiveIngress struct {
	Pos lexer.Position

	Type directiveIngressType `parser:"'ingress' @@"`
}

func (*directiveIngress) directive()       {}
func (d *directiveIngress) String() string { return fmt.Sprintf("ftl:ingress %s", d.Type) }

type directiveModule struct {
	Pos lexer.Position

	Name string `parser:"'module' @Ident"`
}

func (*directiveModule) directive()       {}
func (d *directiveModule) String() string { return "ftl:module" }

//sumtype:decl
type directiveIngressType interface{ directiveIngressType() }

type directiveIngressHTTP struct {
	Pos lexer.Position

	Type   string `parser:"@('http' | 'ftl')?"`
	Method string `parser:"@('GET'|'POST'|'PUT'|'DELETE'|'PATCH')"`
	Path   string `parser:"@(('/' (('{' Ident '}') | Ident))+ '/'?)"`
}

func (d *directiveIngressHTTP) directiveIngressType() {}
func (d *directiveIngressHTTP) String() string {
	typ := d.Type
	if d.Type == "" {
		typ = "ftl"
	}
	return fmt.Sprintf("%s %s %s", typ, d.Method, d.Path)
}

var directiveParser = participle.MustBuild[directiveWrapper](
	participle.Unquote(),
	participle.UseLookahead(2),
	participle.Union[directive](&directiveVerb{}, &directiveIngress{}, &directiveModule{}),
	participle.Union[directiveIngressType](&directiveIngressHTTP{}),
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
