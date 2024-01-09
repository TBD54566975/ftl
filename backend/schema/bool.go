package schema

import (
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

type Bool struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Bool bool `parser:"@'Bool'" protobuf:"-"`
}

var _ Type = (*Bool)(nil)

func (*Bool) schemaChildren() []Node { return nil }
func (*Bool) schemaType()            {}
func (*Bool) String() string         { return "Bool" }

func (b *Bool) ToProto() proto.Message {
	return &schemapb.Bool{Pos: posToProto(b.Pos)}
}
