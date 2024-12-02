package schema

import (
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
)

// Optional represents a Type whose value may be optional.
//
//protobuf:12
type Optional struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Type Type `parser:"@@" protobuf:"2,optional"`
}

var _ Type = (*Optional)(nil)
var _ Symbol = (*Optional)(nil)

func (o *Optional) Equal(other Type) bool {
	ot, ok := other.(*Optional)
	if !ok {
		return false
	}
	return o.Type.Equal(ot.Type)
}
func (o *Optional) Position() Position     { return o.Pos }
func (o *Optional) String() string         { return o.Type.String() + "?" }
func (*Optional) schemaType()              {}
func (*Optional) schemaSymbol()            {}
func (o *Optional) schemaChildren() []Node { return []Node{o.Type} }
func (o *Optional) ToProto() proto.Message {
	return &schemapb.Optional{Pos: posToProto(o.Pos), Type: TypeToProto(o.Type)}
}
