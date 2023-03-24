package schema

import (
	"io"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"google.golang.org/protobuf/proto"
)

var (
	DeclUnion     = []Decl{&Data{}, &Verb{}}
	TypeUnion     = []Type{&Int{}, &Float{}, &String{}, &Bool{}, &Array{}, &Map{}, &VerbRef{}, &DataRef{}}
	MetadataUnion = []Metadata{&MetadataCalls{}}
)

// A Node in the schema grammar.
//
//sumtype:decl
type Node interface {
	String() string
	ToProto() proto.Message
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
	Pos lexer.Position `json:"-" parser:"" protobuf:"-"`

	Int bool `parser:"@'int'" json:"-" protobuf:"-"`
}

type Float struct {
	Pos lexer.Position `json:"-" parser:"" protobuf:"-"`

	Float bool `parser:"@'float'" json:"-" protobuf:"-"`
}

type String struct {
	Pos lexer.Position `json:"-" parser:"" protobuf:"-"`

	Str bool `parser:"@'string'" json:"-" protobuf:"-"`
}

type Bool struct {
	Pos lexer.Position `json:"-" parser:"" protobuf:"-"`

	Bool bool `parser:"@'bool'" json:"-" protobuf:"-"`
}

type Array struct {
	Pos lexer.Position `json:"-" parser:"" protobuf:"-"`

	Element Type `parser:"'[' @@ ']'" json:"element,omitempty" protobuf:"1"`
}

type Map struct {
	Pos lexer.Position `json:"-" parser:"" protobuf:"-"`

	Key   Type `parser:"'{' @@" json:"key,omitempty" protobuf:"1"`
	Value Type `parser:"':' @@ '}'" json:"value,omitempty" protobuf:"2"`
}

type Field struct {
	Pos lexer.Position `json:"-" parser:"" protobuf:"-"`

	Comments []string `parser:"@Comment*" json:"comments,omitempty" protobuf:"2"`
	Name     string   `parser:"@Ident" json:"name,omitempty" protobuf:"1"`
	Type     Type     `parser:"@@" json:"type,omitempty" protobuf:"3"`
}

// Ref is a reference to another symbol.
type Ref struct {
	Pos lexer.Position `json:"-" parser:"" protobuf:"-"`

	Module string `parser:"(@Ident '.')?" json:"module,omitempty" protobuf:"2"`
	Name   string `parser:"@Ident" json:"name,omitempty" protobuf:"1"`
}

// DataRef is a reference to a data structure.
type DataRef Ref

// A Data structure.
type Data struct {
	Pos lexer.Position `json:"-" parser:"" protobuf:"-"`

	Name     string     `parser:"'data' @Ident '{'" json:"name,omitempty" protobuf:"1"`
	Fields   []*Field   `parser:"@@* '}'" json:"fields,omitempty" protobuf:"2"`
	Metadata []Metadata `parser:"@@*" json:"metadata,omitempty" protobuf:"3"`
}

// VerbRef is a reference to a Verb.
type VerbRef Ref

type Verb struct {
	Pos lexer.Position `json:"-" parser:"" protobuf:"-"`

	Comments []string   `parser:"@Comment*" json:"comments,omitempty" protobuf:"2"`
	Name     string     `parser:"'verb' @Ident" json:"name,omitempty" protobuf:"1"`
	Request  *DataRef   `parser:"'(' @@ ')'" json:"request,omitempty" protobuf:"3"`
	Response *DataRef   `parser:"@@" json:"response,omitempty" protobuf:"4"`
	Metadata []Metadata `parser:"@@*" json:"metadata,omitempty" protobuf:"5"`
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
	Pos lexer.Position `json:"-" parser:"" protobuf:"-"`

	Calls []*VerbRef `parser:"'calls' @@ (',' @@)*" json:"calls,omitempty" protobuf:"1"`
}

type Module struct {
	Pos lexer.Position `json:"-" parser:"" protobuf:"-"`

	Comments []string `parser:"@Comment*" json:"comments,omitempty" protobuf:"2"`
	Name     string   `parser:"'module' @Ident '{'" json:"name,omitempty" protobuf:"1"`
	Decls    []Decl   `parser:"@@* '}'" json:"decls,omitempty" protobuf:"3"`
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
	Pos lexer.Position `json:"-" parser:"" protobuf:"-"`

	Modules []*Module `parser:"@@*" json:"modules,omitempty" protobuf:"1"`
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

var (
	lex = lexer.MustSimple([]lexer.SimpleRule{
		{Name: "Whitespace", Pattern: `\s+`},
		{Name: "Ident", Pattern: `\b[a-zA-Z_][a-zA-Z0-9_]*\b`},
		{Name: "Comment", Pattern: `//.*`},
		{Name: "String", Pattern: `"(?:\\.|[^"])*"`},
		{Name: "Number", Pattern: `[0-9]+(?:\.[0-9]+)?`},
		{Name: "Punct", Pattern: `[-:[\]{}<>()*+?.,\\^$|#]`},
	})

	parserOptions = []participle.Option{
		participle.Lexer(lex),
		participle.Elide("Whitespace"),
		participle.Unquote(),
		participle.Map(func(token lexer.Token) (lexer.Token, error) {
			token.Value = strings.TrimSpace(strings.TrimPrefix(token.Value, "//"))
			return token, nil
		}, "Comment"),
		participle.Union(TypeUnion...),
		participle.Union(MetadataUnion...),
		participle.Union(DeclUnion...),
	}

	parser       = participle.MustBuild[Schema](parserOptions...)
	moduleParser = participle.MustBuild[Module](parserOptions...)
)

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
