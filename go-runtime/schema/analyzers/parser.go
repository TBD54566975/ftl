package analyzers

import (
	"errors"
	"fmt"
	"go/ast"
	"strconv"
	"strings"

	"github.com/TBD54566975/golang-tools/go/analysis"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"github.com/TBD54566975/ftl/backend/schema"
)

// This file contains a parser for Go FTL directives.
//
// eg. //ftl:ingress http GET /foo/bar

type directiveWrapper struct {
	Directive directive `parser:"'ftl' ':' @@"`
}

//sumtype:decl
type directive interface{ directive() }

type exportable interface {
	IsExported() bool
}

type directiveVerb struct {
	Pos lexer.Position

	Verb   bool `parser:"@'verb'"`
	Export bool `parser:"@'export'?"`
}

func (*directiveVerb) directive() {}
func (d *directiveVerb) String() string {
	if d.Export {
		return "ftl:verb export"
	}
	return "ftl:verb"
}
func (d *directiveVerb) IsExported() bool {
	return d.Export
}

type directiveData struct {
	Pos lexer.Position

	Data   bool `parser:"@'data'"`
	Export bool `parser:"@'export'?"`
}

func (*directiveData) directive() {}
func (d *directiveData) String() string {
	if d.Export {
		return "ftl:data export"
	}
	return "ftl:data"
}
func (d *directiveData) IsExported() bool {
	return d.Export
}

type directiveEnum struct {
	Pos lexer.Position

	Enum   bool `parser:"@'enum'"`
	Export bool `parser:"@'export'?"`
}

func (*directiveEnum) directive() {}
func (d *directiveEnum) String() string {
	if d.Export {
		return "ftl:enum export"
	}
	return "ftl:enum"
}
func (d *directiveEnum) IsExported() bool {
	return d.Export
}

type directiveTypeAlias struct {
	Pos lexer.Position

	TypeAlias bool `parser:"@'typealias'"`
	Export    bool `parser:"@'export'?"`
}

func (*directiveTypeAlias) directive() {}
func (d *directiveTypeAlias) String() string {
	if d.Export {
		return "ftl:typealias export"
	}
	return "ftl:typealias"
}
func (d *directiveTypeAlias) IsExported() bool {
	return d.Export
}

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

type directiveCronJob struct {
	Pos schema.Position

	Cron string `parser:"'cron' Whitespace @((' ' | Number | '-' | '/' | '*' | ',')+)"`
}

func (*directiveCronJob) directive() {}

func (d *directiveCronJob) String() string {
	return fmt.Sprintf("cron %s", d.Cron)
}

type directiveRetry struct {
	Pos schema.Position

	Count      *int   `parser:"'retry' (@Number Whitespace)?"`
	MinBackoff string `parser:"@(Number (?! Whitespace) Ident)?"`
	MaxBackoff string `parser:"@(Number (?! Whitespace) Ident)?"`
}

func (*directiveRetry) directive() {}

func (d *directiveRetry) String() string {
	components := []string{"retry"}
	if d.Count != nil {
		components = append(components, strconv.Itoa(*d.Count))
	}
	components = append(components, d.MinBackoff)
	if len(d.MaxBackoff) > 0 {
		components = append(components, d.MaxBackoff)
	}
	return strings.Join(components, " ")
}

// used to subscribe a sink to a subscription
type directiveSubscriber struct {
	Pos schema.Position

	Name string `parser:"'subscribe' @Ident"`
}

func (*directiveSubscriber) directive() {}

func (d *directiveSubscriber) String() string {
	return fmt.Sprintf("subscribe %s", d.Name)
}

// most declarations include export in other directives, but some don't have any other way.
type directiveExport struct {
	Pos schema.Position

	Export bool `parser:"@'export'"`
}

func (*directiveExport) directive() {}

func (d *directiveExport) String() string {
	return "export"
}

var directiveParser = participle.MustBuild[directiveWrapper](
	participle.Lexer(schema.Lexer),
	participle.Elide("Whitespace"),
	participle.Unquote(),
	participle.UseLookahead(2),
	participle.Union[directive](&directiveVerb{}, &directiveData{}, &directiveEnum{}, &directiveTypeAlias{},
		&directiveIngress{}, &directiveCronJob{}, &directiveRetry{}, &directiveSubscriber{}, &directiveExport{}),
	participle.Union[schema.IngressPathComponent](&schema.IngressPathLiteral{}, &schema.IngressPathParameter{}),
)

func parseDirectives(pass *analysis.Pass, node ast.Node, docs *ast.CommentGroup) []directive {
	if docs == nil {
		return nil
	}
	directives := []directive{}
	for _, line := range docs.List {
		if !strings.HasPrefix(line.Text, "//ftl:") {
			continue
		}
		pos := pass.Fset.Position(line.Pos())
		// TODO: We need to adjust position information embedded in the schema.
		directive, err := directiveParser.ParseString(pos.Filename, line.Text[2:])
		if err != nil {
			// Adjust the Participle-reported position relative to the AST node.
			var perr participle.Error
			if errors.As(err, &perr) {
				file := pass.Fset.File(node.Pos())
				pass.Report(errorfAtPos(file.Pos(pos.Line), file.Pos(pos.Column+2), "%s", perr.Message()))
			} else {
				pass.Report(wrapf(node, err, ""))
			}
			return nil
		}
		directives = append(directives, directive.Directive)
	}
	return directives
}
