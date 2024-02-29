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
	Type     Type           `parser:"'(' @@ ')'" protobuf:"4"`
	Variants []*EnumVariant `parser:"'{' @@* '}'" protobuf:"5"`
}

var _ Decl = (*Enum)(nil)

func (e *Enum) Position() Position { return e.Pos }

func (e *Enum) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, encodeComments(e.Comments))
	fmt.Fprintf(w, "enum %s(%s) {\n", e.Name, e.Type.String())
	for _, v := range e.Variants {
		fmt.Fprintln(w, indent(v.String()))
	}
	fmt.Fprint(w, "}")
	return w.String()
}
func (*Enum) schemaDecl() {}
func (e *Enum) schemaChildren() []Node {
	children := make([]Node, 1+len(e.Variants))
	children[0] = e.Type
	for i, v := range e.Variants {
		children[1+i] = v
	}
	return children
}
func (e *Enum) ToProto() proto.Message {
	return &schemapb.Enum{
		Pos:      posToProto(e.Pos),
		Comments: e.Comments,
		Name:     e.Name,
		Type:     typeToProto(e.Type),
		Variants: nodeListToProto[*schemapb.EnumVariant](e.Variants),
	}
}

func EnumFromProto(s *schemapb.Enum) *Enum {
	return &Enum{
		Pos:      posFromProto(s.Pos),
		Name:     s.Name,
		Comments: s.Comments,
		Type:     typeToSchema(s.Type),
		Variants: enumVariantListToSchema(s.Variants),
	}
}

type EnumVariant struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Name  string `parser:"@Ident" protobuf:"2"`
	Value Value  `parser:"'(' @@ ')'" protobuf:"3"`
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
	return fmt.Sprintf("%s(%s)", e.Name, e.Value)
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
