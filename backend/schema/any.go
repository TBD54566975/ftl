package schema

import (
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type Any struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Any bool `parser:"@'Any'" protobuf:"-"`
}

var _ Type = (*Any)(nil)
var _ Symbol = (*Any)(nil)

func (a *Any) Position() Position     { return a.Pos }
func (*Any) schemaChildren() []Node   { return nil }
func (*Any) schemaType()              {}
func (*Any) schemaSymbol()            {}
func (*Any) String() string           { return "Any" }
func (a *Any) ToProto() proto.Message { return &schemapb.Any{Pos: posToProto(a.Pos)} }
func (*Any) GetName() string          { return "" }
