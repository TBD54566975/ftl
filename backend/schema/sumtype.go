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
	Name     string   `parser:"'sumtype' @Ident" protobuf:"3"`
	Types    []Type   `parser:"'{' @@* '}'" protobuf:"4"`
}

var _ Decl = (*SumType)(nil)
var _ Symbol = (*SumType)(nil)

func (s *SumType) Position() Position { return s.Pos }

func (s *SumType) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, encodeComments(s.Comments))

	fmt.Fprintf(w, "sumtype %s {\n", s.Name)
	for _, v := range s.Types {
		fmt.Fprintln(w, indent(v.String()))
	}
	fmt.Fprintf(w, "}")
	return w.String()
}

func (s *SumType) schemaDecl() {}
func (*SumType) schemaSymbol() {}
func (s *SumType) schemaChildren() []Node {
	children := make([]Node, len(s.Types))
	for i, v := range s.Types {
		children[i] = v
	}
	return children
}
func (s *SumType) ToProto() proto.Message {
	types := make([]*schemapb.Type, len(s.Types))
	for i, v := range s.Types {
		types[i] = typeToProto(v)
	}
	return &schemapb.SumType{
		Pos:      posToProto(s.Pos),
		Comments: s.Comments,
		Name:     s.Name,
		Types:    types,
		//Types:    nodeListToProto[*schemapb.Type](s.Types),
	}
}

func (s *SumType) GetName() string { return s.Name }

func SumTypeFromProto(s *schemapb.SumType) *SumType {
	types := make([]Type, len(s.Types))
	for i, v := range s.Types {
		types[i] = typeToSchema(v)
	}
	return &SumType{
		Pos:      posFromProto(s.Pos),
		Name:     s.Name,
		Comments: s.Comments,
		Types:    types,
	}
}
