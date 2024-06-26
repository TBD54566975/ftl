package compile

// TODO: This file is now duplicated in go-runtime/schema/analyzers/parser.go. It should be removed
// from here once the schema extraction refactoring is complete.

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/cron"
)

// This file contains a parser for Go FTL directives.
//
// eg. //ftl:ingress http GET /foo/bar

type directiveWrapper struct {
	Directive directive `parser:"'ftl' ':' @@"`
}

//sumtype:decl
type directive interface {
	directive()
	SetPosition(pos schema.Position)
	GetPosition() schema.Position
}

type exportable interface {
	IsExported() bool
}

type directiveVerb struct {
	Pos schema.Position

	Verb   bool `parser:"@'verb'"`
	Export bool `parser:"@'export'?"`
}

func (*directiveVerb) directive() {}

func (d *directiveVerb) SetPosition(pos schema.Position) {
	d.Pos = pos
}

func (d *directiveVerb) GetPosition() schema.Position {
	return d.Pos
}

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
	Pos schema.Position

	Data   bool `parser:"@'data'"`
	Export bool `parser:"@'export'?"`
}

func (*directiveData) directive() {}

func (d *directiveData) SetPosition(pos schema.Position) {
	d.Pos = pos
}

func (d *directiveData) GetPosition() schema.Position {
	return d.Pos
}

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
	Pos schema.Position

	Enum   bool `parser:"@'enum'"`
	Export bool `parser:"@'export'?"`
}

func (*directiveEnum) directive() {}

func (d *directiveEnum) SetPosition(pos schema.Position) {
	d.Pos = pos
}

func (d *directiveEnum) GetPosition() schema.Position {
	return d.Pos
}

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
	Pos schema.Position

	TypeAlias bool `parser:"@'typealias'"`
	Export    bool `parser:"@'export'?"`
}

func (*directiveTypeAlias) directive() {}

func (d *directiveTypeAlias) SetPosition(pos schema.Position) {
	d.Pos = pos
}

func (d *directiveTypeAlias) GetPosition() schema.Position {
	return d.Pos
}

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

func (d *directiveIngress) SetPosition(pos schema.Position) {
	d.Pos = pos
}

func (d *directiveIngress) GetPosition() schema.Position {
	return d.Pos
}

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

	Cron cron.Pattern `parser:"'cron' @@"`
}

func (*directiveCronJob) directive() {}

func (d *directiveCronJob) SetPosition(pos schema.Position) {
	d.Pos = pos
}

func (d *directiveCronJob) GetPosition() schema.Position {
	return d.Pos
}

func (d *directiveCronJob) String() string {
	return fmt.Sprintf("ftl:cron %s", d.Cron)
}

type directiveRetry struct {
	Pos schema.Position

	Count      *int   `parser:"'retry' (@Number Whitespace)?"`
	MinBackoff string `parser:"@(Number (?! Whitespace) Ident)?"`
	MaxBackoff string `parser:"@(Number (?! Whitespace) Ident)?"`
}

func (*directiveRetry) directive() {}

func (d *directiveRetry) SetPosition(pos schema.Position) {
	d.Pos = pos
}

func (d *directiveRetry) GetPosition() schema.Position {
	return d.Pos
}

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

func (d *directiveSubscriber) SetPosition(pos schema.Position) {
	d.Pos = pos
}

func (d *directiveSubscriber) GetPosition() schema.Position {
	return d.Pos
}

func (d *directiveSubscriber) String() string {
	return fmt.Sprintf("subscribe %s", d.Name)
}

// most declarations include export in other directives, but some don't have any other way.
type directiveExport struct {
	Pos schema.Position

	Export bool `parser:"@'export'"`
}

func (*directiveExport) directive() {}

func (d *directiveExport) SetPosition(pos schema.Position) {
	d.Pos = pos
}

func (d *directiveExport) GetPosition() schema.Position {
	return d.Pos
}

func (d *directiveExport) String() string {
	return "ftl:export"
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
		ppos := schema.Position{
			Filename: pos.Filename,
			Line:     pos.Line,
			Column:   pos.Column + 2, // Skip "//"
		}

		directive, err := directiveParser.ParseString(pos.Filename, line.Text[2:])
		if err != nil {
			var scerr *schema.Error
			var perr participle.Error
			if errors.As(err, &perr) {
				scerr = schema.Errorf(ppos, ppos.Column, "%s", perr.Message())
			} else {
				scerr = wrapf(node, err, "")
			}
			return nil, scerr
		}

		directive.Directive.SetPosition(ppos)
		directives = append(directives, directive.Directive)
	}
	return directives, nil
}
