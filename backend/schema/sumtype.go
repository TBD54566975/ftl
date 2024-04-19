package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	//schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type SumType struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" protobuf:"2"`
	Name     string   `parser:"'sumtype' @Ident '='" protobuf:"3"`
	Types    []Type   `parser:"<expr> | <expr> | ..." protobuf:"4"`
}

var _ Decl = (*SumType)(nil)
var _ Symbol = (*SumType)(nil)

func (s *SumType) Position() Position { return s.Pos }

func (s *SumType) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, encodeComments(s.Comments))

	typeNames := make([]string, len(s.Types))
	for i, v := range s.Types {
		typeNames[i] = v.String()
	}
	fmt.Fprintf(w, "sumtype %s = %s", s.Name, strings.Join(typeNames, " | "))
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
	/*protoTypes := make([]schemapb.Type, len(s.Types))
	for i, v := range s.Types {
		protoTypes[i] = *typeToProto(v)
	}
	return &schemapb.SumType{
		Pos:         posToProto(s.Pos),
		Comments:    s.Comments,
		Name:        s.Name,
		Types: protoTypes,
	}*/
	return nil
}

func (s *SumType) GetName() string { return s.Name }

/*func SumTypeFromProto(s *schemapb.SumType) *SumType {
	types := make([]Type, len(s.Types))
	for i, v := range s.Types {
		t := typeToSchema(v)
		types[i] = t
	}
	return &SumType{
		Pos:         posFromProto(s.Pos),
		Name:        s.Name,
		Comments:    s.Comments,
		Types:       types,
	}
}
*/
