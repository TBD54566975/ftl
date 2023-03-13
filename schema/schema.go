package schema

import (
	"io"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// A Node in the schema grammar.
//
//sumtype:decl
type Node interface {
	String() string
	// schemaChildren returns the children of this node.
	schemaChildren() []Node
}

// Type represents a Type Node in the schema grammar.
//
//sumtype:decl
type Type interface {
	Node
	// schemaType is a marker to ensure that all types implement the Type interface.
	schemaType()
}

type Int struct {
	Pos lexer.Position `parser:"" json:"-"`

	Int bool `parser:"@'int'" json:"bool"`
}

type Float struct {
	Pos lexer.Position `parser:"" json:"-"`

	Float bool `parser:"@'float'" json:"float"`
}

type String struct {
	Pos lexer.Position `parser:"" json:"-"`

	Str bool `parser:"@'string'" json:"str"`
}

type Bool struct {
	Pos lexer.Position `parser:"" json:"-"`

	Bool bool `parser:"@'bool'" json:"bool"`
}

type Array struct {
	Pos lexer.Position `parser:"" json:"-"`

	Element Type `parser:"'array' '<' @@ '>'" json:"array"`
}

type Map struct {
	Pos lexer.Position `parser:"" json:"-"`

	Key   Type `parser:"'map' '<' @@" json:"key"`
	Value Type `parser:"',' @@ '>'" json:"value"`
}

type Field struct {
	Pos lexer.Position `parser:"" json:"-"`

	Name string `parser:"@Ident" json:"name"`
	Type Type   `parser:"@@" json:"type"`
}

// Ref is a reference to another symbol.
type Ref struct {
	Pos lexer.Position `parser:"" json:"-"`

	Module string `parser:"(@Ident '.')?" json:"module,omitempty"`
	Name   string `parser:"@Ident" json:"name"`
}

// DataRef is a reference to a data structure.
type DataRef Ref

// A Data structure.
type Data struct {
	Pos lexer.Position `parser:"" json:"-"`

	Name   string  `parser:"'data' @Ident '{'" json:"name"`
	Fields []Field `parser:"@@* '}'" json:"fields,omitempty"`
}

// VerbRef is a reference to a Verb.
type VerbRef Ref

type Verb struct {
	Pos lexer.Position `parser:"" json:"-"`

	Name     string    `parser:"'verb' @Ident" json:"name"`
	Request  DataRef   `parser:"'(' @@ ')'" json:"request"`
	Response DataRef   `parser:"@@" json:"response"`
	Calls    []VerbRef `parser:"('calls' @@ (',' @@)*)?" json:"calls,omitempty"`
}

type Module struct {
	Pos lexer.Position `parser:"" json:"-"`

	Name  string `parser:"'module' @Ident '{'" json:"name"`
	Data  []Data `parser:"@@*" json:"data,omitempty"`
	Verbs []Verb `parser:"@@* '}'" json:"verbs,omitempty"`
}

type Schema struct {
	Pos lexer.Position `parser:"" json:"-"`

	Modules []Module `parser:"@@*"`
}

var parser = participle.MustBuild[Schema](
	participle.UseLookahead(2),
	participle.Union[Type](Int{}, Float{}, String{}, Bool{}, Array{}, Map{}, VerbRef{}, DataRef{}),
)

// Validate performs semantic analysis of the module.
func Validate(schema Schema) error {
	verbs := map[string]bool{}
	data := map[string]bool{}
	verbRefs := []VerbRef{}
	dataRefs := []DataRef{}
	for _, module := range schema.Modules {
		err := Visit(module, func(n Node, next func() error) error {
			switch n := n.(type) {
			case VerbRef:
				verbRefs = append(verbRefs, n)
			case DataRef:
				dataRefs = append(dataRefs, n)
			case Verb:
				verbs[module.Name+"."+n.Name] = true
				verbs[n.Name] = true
			case Data:
				data[module.Name+"."+n.Name] = true
				data[n.Name] = true
			default:
			}
			return next()
		})
		if err != nil {
			return errors.WithStack(err)
		}
	}
	merr := []error{}
	for _, ref := range verbRefs {
		if !verbs[ref.String()] {
			merr = append(merr, errors.Errorf("%s: reference to unknown Verb %q", ref.Pos, ref))
		}
	}
	for _, ref := range dataRefs {
		if !data[ref.String()] {
			merr = append(merr, errors.Errorf("%s: reference to unknown Data structure %q", ref.Pos, ref))
		}
	}
	return errors.Join(merr...)
}

func ParseBytes(filename string, input []byte) (Schema, error) {
	mod, err := parser.ParseBytes(filename, input)
	if err != nil {
		return Schema{}, errors.WithStack(err)
	}
	return *mod, Validate(*mod)
}

func ParseString(filename, input string) (Schema, error) {
	mod, err := parser.ParseString(filename, input)
	if err != nil {
		return Schema{}, errors.WithStack(err)
	}
	return *mod, Validate(*mod)
}

func Parse(filename string, r io.Reader) (Schema, error) {
	mod, err := parser.Parse(filename, r)
	if err != nil {
		return Schema{}, errors.WithStack(err)
	}
	return *mod, Validate(*mod)
}
