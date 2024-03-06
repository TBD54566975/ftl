package schema

import (
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type String struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Str bool `parser:"@'String'" protobuf:"-"`
}

var _ Type = (*String)(nil)
var _ Decl = (*String)(nil)

func (s *String) Position() Position     { return s.Pos }
func (*String) schemaChildren() []Node   { return nil }
func (*String) schemaType()              {}
func (*String) schemaDecl()              {}
func (*String) String() string           { return "String" }
func (s *String) ToProto() proto.Message { return &schemapb.String{Pos: posToProto(s.Pos)} }
func (*String) GetName() string          { return "" }
