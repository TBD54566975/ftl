package schema

import (
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

// Optional represents a Type whose value may be optional.
type Optional struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Type Type `parser:"@@" protobuf:"2,optional"`
}

var _ Type = (*Optional)(nil)

func (o *Optional) Position() Position     { return o.Pos }
func (o *Optional) String() string         { return o.Type.String() + "?" }
func (*Optional) schemaType()              {}
func (o *Optional) schemaChildren() []Node { return []Node{o.Type} }
func (o *Optional) ToProto() proto.Message {
	return &schemapb.Optional{Pos: posToProto(o.Pos), Type: typeToProto(o.Type)}
}
