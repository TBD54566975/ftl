package schema

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"google.golang.org/protobuf/proto"
)

var (
	declUnion            = []Decl{&Data{}, &Verb{}, &Database{}, &Enum{}}
	nonOptionalTypeUnion = []Type{
		&Int{}, &Float{}, &String{}, &Bytes{}, &Bool{}, &Time{}, &Array{},
		&Map{}, &Any{}, &Unit{},
		// Note: any types resolved by identifier (eg. "Any", "Unit", etc.) must
		// be prior to DataRef.
		&DataRef{},
	}
	typeUnion     = append(nonOptionalTypeUnion, &Optional{})
	metadataUnion = []Metadata{&MetadataCalls{}, &MetadataIngress{}, &MetadataDatabases{}, &MetadataAlias{}}
	ingressUnion  = []IngressPathComponent{&IngressPathLiteral{}, &IngressPathParameter{}}
	valueUnion    = []Value{&StringValue{}, &IntValue{}}

	// Used by protobuf generation.
	unions = map[reflect.Type][]reflect.Type{
		reflect.TypeOf((*Type)(nil)).Elem():                 reflectUnion(typeUnion...),
		reflect.TypeOf((*Metadata)(nil)).Elem():             reflectUnion(metadataUnion...),
		reflect.TypeOf((*IngressPathComponent)(nil)).Elem(): reflectUnion(ingressUnion...),
		reflect.TypeOf((*Decl)(nil)).Elem():                 reflectUnion(declUnion...),
		reflect.TypeOf((*Value)(nil)).Elem():                reflectUnion(valueUnion...),
	}

	Lexer = lexer.MustSimple([]lexer.SimpleRule{
		{Name: "Whitespace", Pattern: `\s+`},
		{Name: "Ident", Pattern: `\b[a-zA-Z_][a-zA-Z0-9_]*\b`},
		{Name: "Comment", Pattern: `//.*`},
		{Name: "String", Pattern: `"(?:\\.|[^"])*"`},
		{Name: "Number", Pattern: `[0-9]+(?:\.[0-9]+)?`},
		{Name: "Punct", Pattern: `[%/\-\_:[\]{}<>()*+?.,\\^$|#~!\'@]`},
	})

	commonParserOptions = []participle.Option{
		participle.Lexer(Lexer),
		participle.Elide("Whitespace"),
		participle.Unquote(),
		participle.Map(func(token lexer.Token) (lexer.Token, error) {
			token.Value = strings.TrimSpace(strings.TrimPrefix(token.Value, "//"))
			return token, nil
		}, "Comment"),
		participle.Union(metadataUnion...),
		participle.Union(ingressUnion...),
		participle.Union(declUnion...),
		participle.Union(valueUnion...),
	}

	// Parser options for every parser _except_ the type parser.
	parserOptions = append(commonParserOptions, participle.ParseTypeWith(parseType))

	parser        = participle.MustBuild[Schema](parserOptions...)
	moduleParser  = participle.MustBuild[Module](parserOptions...)
	typeParser    = participle.MustBuild[typeParserGrammar](append(commonParserOptions, participle.Union(nonOptionalTypeUnion...))...)
	dataRefParser = participle.MustBuild[DataRef](parserOptions...)
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

func (p Position) ToProto() proto.Message { return posToProto(p) }

// A Node in the schema grammar.
//
//sumtype:decl
type Node interface {
	String() string
	ToProto() proto.Message
	Position() Position
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

// Metadata represents a metadata Node in the schema grammar.
//
//sumtype:decl
type Metadata interface {
	Node
	schemaMetadata()
}

// Value represents a value Node in the schema grammar.
//
//sumtype:decl
type Value interface {
	Node
	schemaValueType() Type
}

// Decl represents a type declaration in the schema grammar.
//
//sumtype:decl
type Decl interface {
	Node
	schemaDecl()
}

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
	return Validate(mod)
}

func ParseModuleString(filename, input string) (*Module, error) {
	mod, err := moduleParser.ParseString(filename, input)
	if err != nil {
		return nil, err
	}
	return mod, ValidateModule(mod)
}

func Parse(filename string, r io.Reader) (*Schema, error) {
	mod, err := parser.Parse(filename, r)
	if err != nil {
		return nil, err
	}
	return Validate(mod)
}

func ParseModule(filename string, r io.Reader) (*Module, error) {
	mod, err := moduleParser.Parse(filename, r)
	if err != nil {
		return nil, err
	}
	return mod, ValidateModule(mod)
}
