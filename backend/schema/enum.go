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
	Name     string         `parser:"'enum' @Ident" protobuf:"3"`
	Variants []*EnumVariant `parser:"'{' @@* '}'" protobuf:"4"`
}

var _ Decl = (*Enum)(nil)
var _ Symbol = (*Enum)(nil)

func (e *Enum) Position() Position { return e.Pos }

func (e *Enum) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, encodeComments(e.Comments))
	fmt.Fprintf(w, "enum %s {\n", e.Name)
	for _, v := range e.Variants {
		fmt.Fprintln(w, indent(v.String()))
	}
	fmt.Fprint(w, "}")
	return w.String()
}
func (*Enum) schemaDecl()   {}
func (*Enum) schemaSymbol() {}
func (e *Enum) schemaChildren() []Node {
	children := make([]Node, len(e.Variants))
	for i, v := range e.Variants {
		children[i] = v
	}
	return children
}
func (e *Enum) ToProto() proto.Message {
	return &schemapb.Enum{
		Pos:      posToProto(e.Pos),
		Comments: e.Comments,
		Name:     e.Name,
		Variants: nodeListToProto[*schemapb.EnumVariant](e.Variants),
	}
}

func (e *Enum) GetName() string { return e.Name }

func EnumFromProto(s *schemapb.Enum) *Enum {
	return &Enum{
		Pos:      posFromProto(s.Pos),
		Name:     s.Name,
		Comments: s.Comments,
		Variants: enumVariantListToSchema(s.Variants),
	}
}

type EnumVariant struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" protobuf:"2"`
	Name     string   `parser:"@Ident" protobuf:"3"`
	Type     Type     `parser:"@@" protobuf:"4"`
	Value    Value    `parser:"('=' @@)?" protobuf:"5,optional"`
}

func (e *EnumVariant) ToProto() proto.Message {
	return &schemapb.EnumVariant{
		Pos:   posToProto(e.Pos),
		Name:  e.Name,
		Type:  typeToProto(e.Type),
		Value: valueToProto(e.Value),
	}
}

func (e *EnumVariant) Position() Position { return e.Pos }

func (e *EnumVariant) schemaChildren() []Node {
	c := []Node{e.Type}
	if e.Value != nil {
		c = append(c, e.Value)
	}
	return c
}

func (e *EnumVariant) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, encodeComments(e.Comments))
	fmt.Fprintf(w, e.Name)
	fmt.Fprintf(w, " %s", e.Type)
	if e.Value != nil {
		fmt.Fprint(w, " = ", e.Value)
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
		Type:  typeToSchema(v.Type),
		Value: valueToSchema(v.Value),
	}
}
