package schema

import (
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

type Int struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Int bool `parser:"@'Int'" protobuf:"-"`
}

var _ Type = (*Int)(nil)

func (*Int) schemaChildren() []Node   { return nil }
func (*Int) schemaType()              {}
func (*Int) String() string           { return "Int" }
func (i *Int) ToProto() proto.Message { return &schemapb.Int{Pos: posToProto(i.Pos)} }
