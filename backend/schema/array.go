package schema

import (
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

type Array struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Element Type `parser:"'[' @@ ']'" protobuf:"2"`
}

var _ Type = (*Array)(nil)

func (a *Array) schemaChildren() []Node { return []Node{a.Element} }
func (*Array) schemaType()              {}
func (a *Array) String() string         { return "[" + a.Element.String() + "]" }

func (a *Array) ToProto() proto.Message {
	return &schemapb.Array{Element: typeToProto(a.Element)}
}

func arrayToSchema(s *schemapb.Array) *Array {
	return &Array{
		Pos:     posFromProto(s.Pos),
		Element: typeToSchema(s.Element),
	}
}
