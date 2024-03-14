package schema

import (
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type Time struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Time bool `parser:"@'Time'" protobuf:"-"`
}

var _ Type = (*Time)(nil)
var _ Symbol = (*Time)(nil)

func (t *Time) Position() Position     { return t.Pos }
func (*Time) schemaChildren() []Node   { return nil }
func (*Time) schemaType()              {}
func (*Time) schemaSymbol()            {}
func (*Time) String() string           { return "Time" }
func (t *Time) ToProto() proto.Message { return &schemapb.Time{Pos: posToProto(t.Pos)} }
func (*Time) GetName() string          { return "" }
