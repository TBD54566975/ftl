package schema

import (
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

type Float struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Float bool `parser:"@'Float'" protobuf:"-"`
}

var _ Type = (*Float)(nil)

func (*Float) schemaChildren() []Node   { return nil }
func (*Float) schemaType()              {}
func (*Float) String() string           { return "Float" }
func (f *Float) ToProto() proto.Message { return &schemapb.Float{Pos: posToProto(f.Pos)} }
