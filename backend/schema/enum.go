package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type Enum struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string       `parser:"@Comment*" protobuf:"2"`
	Export   bool           `parser:"@'export'?" protobuf:"3"`
	Name     string         `parser:"'enum' @Ident" protobuf:"4"`
	Type     Type           `parser:"(':' @@)?" protobuf:"5,optional"`
	Variants []*EnumVariant `parser:"'{' @@* '}'" protobuf:"6"`
}

var _ Decl = (*Enum)(nil)
var _ Symbol = (*Enum)(nil)

func (e *Enum) Position() Position { return e.Pos }

func (e *Enum) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(e.Comments))
	if e.Export {
		fmt.Fprint(w, "export ")
	}
	fmt.Fprintf(w, "enum %s", e.Name)
	if e.Type != nil {
		fmt.Fprintf(w, ": %s", e.Type)
	}
	fmt.Fprint(w, " {\n")
	for _, v := range e.Variants {
		fmt.Fprintln(w, indent(v.String()))
	}
	fmt.Fprint(w, "}")
	return w.String()
}
func (*Enum) schemaDecl()   {}
func (*Enum) schemaSymbol() {}
func (e *Enum) schemaChildren() []Node {
	var children []Node
	for _, v := range e.Variants {
		children = append(children, v)
	}
	if e.Type != nil {
		children = append(children, e.Type)
	}
	return children
}
func (e *Enum) ToProto() proto.Message {
	se := &schemapb.Enum{
		Pos:      posToProto(e.Pos),
		Comments: e.Comments,
		Name:     e.Name,
		Export:   e.Export,
		Variants: nodeListToProto[*schemapb.EnumVariant](e.Variants),
	}
	if e.Type != nil {
		se.Type = typeToProto(e.Type)
	}
	return se
}

func (e *Enum) GetName() string  { return e.Name }
func (e *Enum) IsExported() bool { return e.Export }

func EnumFromProto(s *schemapb.Enum) *Enum {
	e := &Enum{
		Pos:      posFromProto(s.Pos),
		Name:     s.Name,
		Export:   s.Export,
		Comments: s.Comments,
		Variants: enumVariantListToSchema(s.Variants),
	}
	if s.Type != nil {
		e.Type = typeToSchema(s.Type)
	}
	return e
}

// IsValueEnum determines whether this is a type or value enum using `e.Type` alone
// because value enums must always have a unified type across all variants, whereas type
// enums by definition cannot have a unified type.
func (e *Enum) IsValueEnum() bool {
	return e.Type != nil
}

func (e *Enum) VariantForName(name string) (*EnumVariant, bool) {
	for _, v := range e.Variants {
		if name == v.Name {
			return v, true
		}
	}
	return nil, false
}

type EnumVariant struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" protobuf:"2"`
	Name     string   `parser:"@Ident" protobuf:"3"`
	Value    Value    `parser:"(('=' @@) | @@)!" protobuf:"4"`
}

func (e *EnumVariant) ToProto() proto.Message {
	return &schemapb.EnumVariant{
		Pos:   posToProto(e.Pos),
		Name:  e.Name,
		Value: valueToProto(e.Value),
	}
}

func (e *EnumVariant) Position() Position { return e.Pos }

func (e *EnumVariant) schemaChildren() []Node { return []Node{e.Value} }

func (e *EnumVariant) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(e.Comments))
	fmt.Fprintf(w, e.Name)
	if e.Value != nil {
		if _, ok := e.Value.(*TypeValue); ok {
			fmt.Fprint(w, " ", e.Value)
		} else {
			fmt.Fprint(w, " = ", e.Value)
		}
	}
	return w.String()
}

func enumVariantListToSchema(e []*schemapb.EnumVariant) []*EnumVariant {
	out := make([]*EnumVariant, 0, len(e))
	for _, v := range e {
		out = append(out, enumVariantToSchema(v))
	}
	return out
}

func enumVariantToSchema(v *schemapb.EnumVariant) *EnumVariant {
	return &EnumVariant{
		Pos:   posFromProto(v.Pos),
		Name:  v.Name,
		Value: valueToSchema(v.Value),
	}
}
