package schema

import (
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

//protobuf:7
type Array struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Element Type `parser:"'[' @@ ']'" protobuf:"2"`
}

var _ Type = (*Array)(nil)
var _ Symbol = (*Array)(nil)

func (a *Array) Equal(other Type) bool {
	o, ok := other.(*Array)
	if !ok {
		return false
	}
	return a.Element.Equal(o.Element)
}
func (a *Array) Position() Position     { return a.Pos }
func (a *Array) schemaChildren() []Node { return []Node{a.Element} }
func (a *Array) schemaType()            {}
func (a *Array) schemaSymbol()          {}
func (a *Array) String() string         { return "[" + a.Element.String() + "]" }

func (a *Array) ToProto() proto.Message {
	return &schemapb.Array{Element: TypeToProto(a.Element)}
}

func arrayToSchema(s *schemapb.Array) *Array {
	return &Array{
		Pos:     PosFromProto(s.Pos),
		Element: TypeFromProto(s.Element),
	}
}
