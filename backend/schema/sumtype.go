package schema

import (
	"fmt"
	"strings"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"google.golang.org/protobuf/proto"
)

type SumType struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string          `parser:"@Comment*" protobuf:"2"`
	Name     string            `parser:"'sumtype' @Ident '='" protobuf:"3"`
	Variants []*SumTypeVariant `parser:"@@*" protobuf:"4"`
}

var _ Decl = (*SumType)(nil)
var _ Symbol = (*SumType)(nil)

func (s *SumType) Position() Position { return s.Pos }

func (s *SumType) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, encodeComments(s.Comments))

	typeNames := make([]string, len(s.Variants))
	for i, v := range s.Variants {
		typeNames[i] = v.String()
	}
	fmt.Fprintf(w, "sumtype %s = %s", s.Name, strings.Join(typeNames, " | "))
	return w.String()
}

func (s *SumType) schemaDecl() {}
func (*SumType) schemaSymbol() {}
func (s *SumType) schemaChildren() []Node {
	children := make([]Node, len(s.Variants))
	for i, v := range s.Variants {
		children[i] = v
	}
	return children
}
func (s *SumType) ToProto() proto.Message {
	return &schemapb.SumType{
		Pos:      posToProto(s.Pos),
		Comments: s.Comments,
		Name:     s.Name,
		Variants: nodeListToProto[*schemapb.SumTypeVariant](s.Variants),
	}
}

func (s *SumType) GetName() string { return s.Name }

func SumTypeFromProto(s *schemapb.SumType) *SumType {
	variants := make([]*SumTypeVariant, len(s.Variants))
	for i, v := range s.Variants {
		t := sumTypeVariantFromProto(v)
		variants[i] = t
	}
	return &SumType{
		Pos:      posFromProto(s.Pos),
		Name:     s.Name,
		Comments: s.Comments,
		Variants: variants,
	}
}

type SumTypeVariant struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	// Question about lookahead buffers?
	Type Type `parser:"@@ '|'?" protobuf:"2"`
}

func (s *SumTypeVariant) ToProto() proto.Message {
	return &schemapb.SumTypeVariant{
		Pos:  posToProto(s.Pos),
		Type: typeToProto(s.Type),
	}
}

func (s *SumTypeVariant) Position() Position { return s.Pos }

func (s *SumTypeVariant) schemaChildren() []Node { return []Node{s.Type} }

func (s *SumTypeVariant) String() string { return s.Type.String() }

func sumTypeVariantFromProto(v *schemapb.SumTypeVariant) *SumTypeVariant {
	return &SumTypeVariant{
		Pos:  posFromProto(v.Pos),
		Type: typeToSchema(v.Type),
	}
}
