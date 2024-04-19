package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	//schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type SumType struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string          `parser:"@Comment*" protobuf:"2"`
	Name     string            `parser:"'sumtype' @Ident '='" protobuf:"3"`
	Variants []*SumTypeVariant `parser:"@@" protobuf:"4"`
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
	/*protoTypes := make([]schemapb.Type, len(s.Variants))
	for i, v := range s.Variants {
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
	types := make([]Type, len(s.Variants))
	for i, v := range s.Variants {
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

type SumTypeVariant struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Type Type `parser:"@@ '|'?" protobuf:"2"`
}

func (s *SumTypeVariant) ToProto() proto.Message {
	/*return &schemapb.EnumVariant{
		Pos:  posToProto(e.Pos),
		Type: typeToProto(v),
	}*/
	return nil
}

func (s *SumTypeVariant) Position() Position { return s.Pos }

func (s *SumTypeVariant) schemaChildren() []Node { return []Node{s.Type} }

func (s *SumTypeVariant) String() string { return s.Type.String() }

// todo pb converters
