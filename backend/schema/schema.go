package schema

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
)

var (
	declUnion            = []Decl{&Data{}, &Verb{}}
	nonOptionalTypeUnion = []Type{&Int{}, &Float{}, &String{}, &Bool{}, &Time{}, &Array{}, &Map{} /*&VerbRef{},*/, &DataRef{}}
	typeUnion            = append(nonOptionalTypeUnion, &Optional{})
	metadataUnion        = []Metadata{&MetadataCalls{}, &MetadataIngress{}}
	ingressUnion         = []IngressPathComponent{&IngressPathLiteral{}, &IngressPathParameter{}}

	// Used by protobuf generation.
	unions = map[reflect.Type][]reflect.Type{
		reflect.TypeOf((*Type)(nil)).Elem():                 reflectUnion(typeUnion...),
		reflect.TypeOf((*Metadata)(nil)).Elem():             reflectUnion(metadataUnion...),
		reflect.TypeOf((*IngressPathComponent)(nil)).Elem(): reflectUnion(ingressUnion...),
		reflect.TypeOf((*Decl)(nil)).Elem():                 reflectUnion(declUnion...),
	}
)

type Position struct {
	Filename string `protobuf:"1"`
	Offset   int    `parser:"" protobuf:"-"`
	Line     int    `protobuf:"2"`
	Column   int    `protobuf:"3"`
}

func (p Position) String() string {
	if p.Filename == "" {
		return fmt.Sprintf("%d:%d", p.Line, p.Column)
	}
	return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
}

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
	// schemaType is a marker to ensure that all sqltypes implement the Type interface.
	schemaType()
}

// Optional represents a Type whose value may be optional.
type Optional struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Type Type `parser:"@@" protobuf:"2,optional"`
}

type Int struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Int bool `parser:"@'Int'" protobuf:"-"`
}

type Float struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Float bool `parser:"@'Float'" protobuf:"-"`
}

type String struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Str bool `parser:"@'String'" protobuf:"-"`
}

type Bool struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Bool bool `parser:"@'Bool'" protobuf:"-"`
}

type Time struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Time bool `parser:"@'Time'" protobuf:"-"`
}

type Array struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Element Type `parser:"'[' @@ ']'" protobuf:"2"`
}

type Map struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Key   Type `parser:"'{' @@" protobuf:"2"`
	Value Type `parser:"':' @@ '}'" protobuf:"3"`
}

type Field struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" protobuf:"3"`
	Name     string   `parser:"@Ident" protobuf:"2"`
	Type     Type     `parser:"@@" protobuf:"4"`
}

// Ref is a reference to another symbol.
type Ref struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Module string `parser:"(@Ident '.')?" protobuf:"3"`
	Name   string `parser:"@Ident" protobuf:"2"`
}

func (r *Ref) String() string {
	return makeRef(r.Module, r.Name)
}

// DataRef is a reference to a data structure.
type DataRef Ref

// A Data structure.
type Data struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string   `parser:"@Comment*" protobuf:"5"`
	Name     string     `parser:"'data' @Ident '{'" protobuf:"2"`
	Fields   []*Field   `parser:"@@* '}'" protobuf:"3"`
	Metadata []Metadata `parser:"@@*" protobuf:"4"`
}

// VerbRef is a reference to a Verb.
type VerbRef Ref

type Verb struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string   `parser:"@Comment*" protobuf:"3"`
	Name     string     `parser:"'verb' @Ident" protobuf:"2"`
	Request  *DataRef   `parser:"'(' @@ ')'" protobuf:"4"`
	Response *DataRef   `parser:"@@" protobuf:"5"`
	Metadata []Metadata `parser:"@@*" protobuf:"6"`
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
	Pos Position `parser:"" protobuf:"1,optional"`

	Calls []*VerbRef `parser:"'calls' @@ (',' @@)*" protobuf:"2"`
}

type MetadataIngress struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Method string                 `parser:"'ingress' @('GET' | 'POST')" protobuf:"2"`
	Path   []IngressPathComponent `parser:"('/' @@)+" protobuf:"3"`
}

type IngressPathComponent interface {
	Node
	schemaIngressPathComponent()
}

type IngressPathLiteral struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Text string `parser:"@Ident" protobuf:"2"`
}

type IngressPathParameter struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Name string `parser:"'{' @Ident '}'" protobuf:"2"`
}

type Module struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" protobuf:"3"`
	Name     string   `parser:"'module' @Ident '{'" protobuf:"2"`
	Decls    []Decl   `parser:"@@* '}'" protobuf:"4"`
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

func (m *Module) Verbs() []*Verb {
	var verbs []*Verb
	for _, d := range m.Decls {
		if v, ok := d.(*Verb); ok {
			verbs = append(verbs, v)
		}
	}
	return verbs
}

func (m *Module) Data() []*Data {
	var data []*Data
	for _, d := range m.Decls {
		if v, ok := d.(*Data); ok {
			data = append(data, v)
		}
	}
	return data
}

func (s *Schema) ResolveDataRef(ref *DataRef) *Data {
	for _, module := range s.Modules {
		if module.Name == ref.Module {
			for _, decl := range module.Decls {
				if data, ok := decl.(*Data); ok && data.Name == ref.Name {
					return data
				}
			}
		}
	}
	return nil
}

func (s *Schema) ResolveVerbRef(ref *VerbRef) *Verb {
	for _, module := range s.Modules {
		if module.Name == ref.Module {
			for _, decl := range module.Decls {
				if verb, ok := decl.(*Verb); ok && verb.Name == ref.Name {
					return verb
				}
			}
		}
	}
	return nil
}

// Imports returns the modules imported by this module.
func (m *Module) Imports() []string {
	imports := map[string]bool{}
	_ = Visit(m, func(n Node, next func() error) error {
		switch n := n.(type) {
		case *DataRef:
			if n.Module != "" && n.Module != m.Name {
				imports[n.Module] = true
			}

		case *VerbRef:
			if n.Module != "" && n.Module != m.Name {
				imports[n.Module] = true
			}

		default:
		}
		return next()
	})

	importStrs := maps.Keys(imports)
	sort.Strings(importStrs)
	return importStrs
}

type Schema struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Modules []*Module `parser:"@@*" protobuf:"2"`
}

func (s *Schema) DataMap() map[DataRef]*Data {
	dataTypes := map[DataRef]*Data{}
	for _, module := range s.Modules {
		for _, decl := range module.Decls {
			if data, ok := decl.(*Data); ok {
				dataTypes[DataRef{Module: module.Name, Name: data.Name}] = data
			}
		}
	}
	return dataTypes
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
		{Name: "Punct", Pattern: `[/-:[\]{}<>()*+?.,\\^$|#]`},
	})

	commonParserOptions = []participle.Option{
		participle.Lexer(lex),
		participle.Elide("Whitespace"),
		participle.Unquote(),
		participle.Map(func(token lexer.Token) (lexer.Token, error) {
			token.Value = strings.TrimSpace(strings.TrimPrefix(token.Value, "//"))
			return token, nil
		}, "Comment"),
		participle.Union(metadataUnion...),
		participle.Union(ingressUnion...),
		participle.Union(declUnion...),
	}

	// Parser options for every parser _except_ the type parser.
	parserOptions = append(commonParserOptions, participle.ParseTypeWith(parseType))

	parser       = participle.MustBuild[Schema](parserOptions...)
	moduleParser = participle.MustBuild[Module](parserOptions...)
	refParser    = participle.MustBuild[Ref](parserOptions...)
	typeParser   = participle.MustBuild[typeParserGrammar](append(commonParserOptions, participle.Union(nonOptionalTypeUnion...))...)
)

// We have a separate parser for types because Participle doesn't support left
// recursion and "Type = Type ? | Int | String ..." is left recursive.
type typeParserGrammar struct {
	Type     Type `parser:"@@"`
	Optional bool `parser:"@'?'?"`
}

func parseType(pl *lexer.PeekingLexer) (Type, error) {
	typ, err := typeParser.ParseFromLexer(pl, participle.AllowTrailing(true))
	if err != nil {
		return nil, err
	}
	if typ.Optional {
		return &Optional{Type: typ.Type}, nil
	}
	return typ.Type, nil
}

func ParseString(filename, input string) (*Schema, error) {
	mod, err := parser.ParseString(filename, input)
	if err != nil {
		return nil, err
	}
	return mod, Validate(mod)
}

func ParseModuleString(filename, input string) (*Module, error) {
	mod, err := moduleParser.ParseString(filename, input)
	if err != nil {
		return nil, err
	}
	return mod, ValidateModule(mod)
}

func ParseRef(ref string) (*Ref, error) {
	r, err := refParser.ParseString("", ref)
	return r, err
}

func Parse(filename string, r io.Reader) (*Schema, error) {
	mod, err := parser.Parse(filename, r)
	if err != nil {
		return nil, err
	}
	return mod, Validate(mod)
}

func ParseModule(filename string, r io.Reader) (*Module, error) {
	mod, err := moduleParser.Parse(filename, r)
	if err != nil {
		return nil, err
	}
	return mod, ValidateModule(mod)
}
