package schema

import (
	"fmt"
	"strings"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"google.golang.org/protobuf/proto"
)

type SumType struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" protobuf:"2"`
	Name     string   `parser:"'sumtype' @Ident Equals" protobuf:"3"`
	Variants []Type   `parser:"@@ ('|' @@)*" protobuf:"4"`
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
	variants := make([]*schemapb.Type, len(s.Variants))
	for i, v := range s.Variants {
		variants[i] = typeToProto(v)
	}
	return &schemapb.SumType{
		Pos:      posToProto(s.Pos),
		Comments: s.Comments,
		Name:     s.Name,
		Variants: variants,
	}
}

func (s *SumType) GetName() string { return s.Name }

func SumTypeFromProto(s *schemapb.SumType) *SumType {
	variants := make([]Type, len(s.Variants))
	for i, v := range s.Variants {
		variants[i] = typeToSchema(v)
	}
	return &SumType{
		Pos:      posFromProto(s.Pos),
		Name:     s.Name,
		Comments: s.Comments,
		Variants: variants,
	}
}
