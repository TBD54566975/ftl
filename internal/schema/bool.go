package schema

import (
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

//protobuf:5
type Bool struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Bool bool `parser:"@'Bool'" protobuf:"-"`
}

var _ Type = (*Bool)(nil)
var _ Symbol = (*Bool)(nil)

func (b *Bool) Equal(other Type) bool  { _, ok := other.(*Bool); return ok }
func (b *Bool) Position() Position     { return b.Pos }
func (*Bool) schemaChildren() []Node   { return nil }
func (*Bool) schemaType()              {}
func (*Bool) schemaSymbol()            {}
func (*Bool) String() string           { return "Bool" }
func (b *Bool) ToProto() proto.Message { return &schemapb.Bool{Pos: posToProto(b.Pos)} }
func (*Bool) GetName() string          { return "Bool" }
