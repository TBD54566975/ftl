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
var _ Decl = (*Bytes)(nil)

func (b *Bytes) Position() Position     { return b.Pos }
func (*Bytes) schemaChildren() []Node   { return nil }
func (*Bytes) schemaType()              {}
func (*Bytes) schemaDecl()              {}
func (*Bytes) String() string           { return "Bytes" }
func (b *Bytes) ToProto() proto.Message { return &schemapb.Bytes{Pos: posToProto(b.Pos)} }
