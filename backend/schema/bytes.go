package schema

import (
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

type Bytes struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Bytes bool `parser:"@'Bytes'" protobuf:"-"`
}

var _ Type = (*Bytes)(nil)

func (*Bytes) schemaChildren() []Node   { return nil }
func (*Bytes) schemaType()              {}
func (*Bytes) String() string           { return "Bytes" }
func (s *Bytes) ToProto() proto.Message { return &schemapb.Bytes{Pos: posToProto(s.Pos)} }
