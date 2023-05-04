package schema

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"google.golang.org/protobuf/proto"
)

var (
	declUnion     = []Decl{&Data{}, &Verb{}}
	typeUnion     = []Type{&Int{}, &Float{}, &String{}, &Bool{}, &Time{}, &Array{}, &Map{}, &VerbRef{}, &DataRef{}}
	metadataUnion = []Metadata{&MetadataCalls{}}

	// Used by protobuf generation.
	unions = map[reflect.Type][]reflect.Type{
		reflect.TypeOf((*Type)(nil)).Elem():     reflectUnion(typeUnion...),
		reflect.TypeOf((*Metadata)(nil)).Elem(): reflectUnion(metadataUnion...),
		reflect.TypeOf((*Decl)(nil)).Elem():     reflectUnion(declUnion...),
	}
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

type Position struct {
	Filename string `json:"filename,omitempty" protobuf:"1"`
	Offset   int    `json:"-" parser:"" protobuf:"-"`
	Line     int    `json:"line,omitempty" protobuf:"2"`
	Column   int    `json:"column,omitempty" protobuf:"3"`
}

func (p Position) String() string {
	if p.Filename == "" {
		return fmt.Sprintf("%d:%d", p.Line, p.Column)
	}
	return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
}

type Int struct {
	Pos Position `json:"-" parser:"" protobuf:"-"`

	Int bool `parser:"@'Int'" json:"-" protobuf:"-"`
}

type Float struct {
	Pos Position `json:"-" parser:"" protobuf:"-"`

	Float bool `parser:"@'Float'" json:"-" protobuf:"-"`
}

type String struct {
	Pos Position `json:"-" parser:"" protobuf:"-"`

	Str bool `parser:"@'String'" json:"-" protobuf:"-"`
}

type Bool struct {
	Pos Position `json:"-" parser:"" protobuf:"-"`

	Bool bool `parser:"@'Bool'" json:"-" protobuf:"-"`
}

type Time struct {
	Pos Position `json:"-" parser:"" protobuf:"-"`

	Time bool `parser:"@'Time'" json:"-" protobuf:"-"`
}

type Array struct {
	Pos Position `json:"-" parser:"" protobuf:"-"`

	Element Type `parser:"'[' @@ ']'" json:"element,omitempty" protobuf:"1"`
}

type Map struct {
	Pos Position `json:"-" parser:"" protobuf:"-"`

	Key   Type `parser:"'{' @@" json:"key,omitempty" protobuf:"1"`
	Value Type `parser:"':' @@ '}'" json:"value,omitempty" protobuf:"2"`
}

type Field struct {
	Pos Position `json:"pos,omitempty" parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" json:"comments,omitempty" protobuf:"3"`
	Name     string   `parser:"@Ident" json:"name,omitempty" protobuf:"2"`
	Type     Type     `parser:"@@" json:"type,omitempty" protobuf:"4"`
}

// Ref is a reference to another symbol.
type Ref struct {
	Pos Position `json:"pos,omitempty" parser:"" protobuf:"1,optional"`

	Module string `parser:"(@Ident '.')?" json:"module,omitempty" protobuf:"3"`
	Name   string `parser:"@Ident" json:"name,omitempty" protobuf:"2"`
}

// DataRef is a reference to a data structure.
type DataRef Ref

// A Data structure.
type Data struct {
	Pos Position `json:"pos,omitempty" parser:"" protobuf:"1,optional"`

	Comments []string   `parser:"@Comment*" json:"comments,omitempty" protobuf:"5"`
	Name     string     `parser:"'data' @Ident '{'" json:"name,omitempty" protobuf:"2"`
	Fields   []*Field   `parser:"@@* '}'" json:"fields,omitempty" protobuf:"3"`
	Metadata []Metadata `parser:"@@*" json:"metadata,omitempty" protobuf:"4"`
}

// VerbRef is a reference to a Verb.
type VerbRef Ref

type Verb struct {
	Pos Position `json:"pos,omitempty" parser:"" protobuf:"1,optional"`

	Comments []string   `parser:"@Comment*" json:"comments,omitempty" protobuf:"3"`
	Name     string     `parser:"'verb' @Ident" json:"name,omitempty" protobuf:"2"`
	Request  *DataRef   `parser:"'(' @@ ')'" json:"request,omitempty" protobuf:"4"`
	Response *DataRef   `parser:"@@" json:"response,omitempty" protobuf:"5"`
	Metadata []Metadata `parser:"@@*" json:"metadata,omitempty" protobuf:"6"`
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
	Pos Position `json:"pos,omitempty" parser:"" protobuf:"1,optional"`

	Calls []*VerbRef `parser:"'calls' @@ (',' @@)*" json:"calls,omitempty" protobuf:"2"`
}

type Module struct {
	Pos Position `json:"pos,omitempty" parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" json:"comments,omitempty" protobuf:"3"`
	Name     string   `parser:"'module' @Ident '{'" json:"name,omitempty" protobuf:"2"`
	Decls    []Decl   `parser:"@@* '}'" json:"decls,omitempty" protobuf:"4"`
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
	Pos Position `json:"pos,omitempty" parser:"" protobuf:"1,optional"`

	Modules []*Module `parser:"@@*" json:"modules,omitempty" protobuf:"2"`
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
		participle.Union(typeUnion...),
		participle.Union(metadataUnion...),
		participle.Union(declUnion...),
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
