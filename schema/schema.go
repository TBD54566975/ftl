package schema

import (
	"io"
	"strings"

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
	Pos lexer.Position

	Int bool `parser:"@'int'"`
}

type Float struct {
	Pos lexer.Position

	Float bool `parser:"@'float'"`
}

type String struct {
	Pos lexer.Position

	Str bool `parser:"@'string'"`
}

type Bool struct {
	Pos lexer.Position

	Bool bool `parser:"@'bool'"`
}

type Array struct {
	Pos lexer.Position

	Element Type `parser:"'[' @@ ']'"`
}

type Map struct {
	Pos lexer.Position

	Key   Type `parser:"'{' @@"`
	Value Type `parser:"':' @@ '}'"`
}

type Field struct {
	Pos lexer.Position

	Comments []string `parser:"@Comment*"`
	Name     string   `parser:"@Ident"`
	Type     Type     `parser:"@@"`
}

// Ref is a reference to another symbol.
type Ref struct {
	Pos lexer.Position

	Module string `parser:"(@Ident '.')?"`
	Name   string `parser:"@Ident"`
}

// DataRef is a reference to a data structure.
type DataRef Ref

// A Data structure.
type Data struct {
	Pos lexer.Position

	Name     string     `parser:"'data' @Ident '{'"`
	Fields   []*Field   `parser:"@@* '}'"`
	Metadata []Metadata `parser:"@@*"`
}

// VerbRef is a reference to a Verb.
type VerbRef Ref

type Verb struct {
	Pos lexer.Position

	Comments []string   `parser:"@Comment*"`
	Name     string     `parser:"'verb' @Ident"`
	Request  *DataRef   `parser:"'(' @@ ')'"`
	Response *DataRef   `parser:"@@"`
	Metadata []Metadata `parser:"@@*"`
}

// AddCall adds a call reference to the Verb.
func (v *Verb) AddCall(verb *VerbRef) {
	for _, c := range v.Metadata {
		if c, ok := c.(*MetadataCalls); ok {
			c.Calls = append(c.Calls, verb)
			return
		}
	}
	v.Metadata = append(v.Metadata, &MetadataCalls{Calls: []*VerbRef{verb}})
}

type Metadata interface {
	Node
	schemaMetadata()
}

type MetadataCalls struct {
	Pos lexer.Position

	Calls []*VerbRef `parser:"'calls' @@ (',' @@)*"`
}

type Module struct {
	Pos lexer.Position

	Comments []string `parser:"@Comment*"`
	Name     string   `parser:"'module' @Ident '{'"`
	Decls    []Decl   `parser:"@@* '}'"`
}

type Decl interface {
	Node
	schemaDecl()
}

// AddData and return its index.
func (m *Module) AddData(data *Data) int {
	for i, d := range m.Decls {
		if d, ok := d.(*Data); ok && d.Name == data.Name {
			return i
		}
	}
	m.Decls = append(m.Decls, data)
	return len(m.Decls) - 1
}

type Schema struct {
	Pos lexer.Position

	Modules []*Module `parser:"@@*"`
}

// Upsert inserts or replaces a module.
func (s *Schema) Upsert(module *Module) {
	for i, m := range s.Modules {
		if m.Name == module.Name {
			s.Modules[i] = module
			return
		}
	}
	s.Modules = append(s.Modules, module)
}

var lex = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Whitespace", Pattern: `\s+`},
	{Name: "Ident", Pattern: `\b[a-zA-Z_][a-zA-Z0-9_]*\b`},
	{Name: "Comment", Pattern: `//.*`},
	{Name: "String", Pattern: `"(?:\\.|[^"])*"`},
	{Name: "Number", Pattern: `[0-9]+(?:\.[0-9]+)?`},
	{Name: "Punct", Pattern: `[-:[\]{}<>()*+?.,\\^$|#]`},
})

var parserOptions = []participle.Option{
	participle.Lexer(lex),
	participle.Elide("Whitespace"),
	participle.Unquote(),
	participle.Map(func(token lexer.Token) (lexer.Token, error) {
		token.Value = strings.TrimSpace(strings.TrimPrefix(token.Value, "//"))
		return token, nil
	}, "Comment"),
	participle.Union[Type](&Int{}, &Float{}, &String{}, &Bool{}, &Array{}, &Map{}, &VerbRef{}, &DataRef{}),
	participle.Union[Metadata](&MetadataCalls{}),
	participle.Union[Decl](&Data{}, &Verb{}),
}

var parser = participle.MustBuild[Schema](parserOptions...)
var moduleParser = participle.MustBuild[Module](parserOptions...)

func ParseString(filename, input string) (*Schema, error) {
	mod, err := parser.ParseString(filename, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return mod, Validate(mod)
}

func ParseModuleString(filename, input string) (*Module, error) {
	mod, err := moduleParser.ParseString(filename, input)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return mod, ValidateModule(mod)
}

func Parse(filename string, r io.Reader) (*Schema, error) {
	mod, err := parser.Parse(filename, r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return mod, Validate(mod)
}

func ParseModule(filename string, r io.Reader) (*Module, error) {
	mod, err := moduleParser.Parse(filename, r)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return mod, ValidateModule(mod)
}
