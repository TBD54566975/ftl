package schema

import (
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

// DataRef is a reference to a data structure.
type DataRef Ref

var _ Type = (*DataRef)(nil)

func (*DataRef) schemaChildren() []Node { return nil }
func (*DataRef) schemaType()            {}
func (s DataRef) String() string        { return makeRef(s.Module, s.Name) }

func (s *DataRef) ToProto() proto.Message {
	return &schemapb.DataRef{
		Pos:    posToProto(s.Pos),
		Name:   s.Name,
		Module: s.Module,
	}
}

func dataRefToSchema(s *schemapb.DataRef) *DataRef {
	return &DataRef{
		Pos:    posFromProto(s.Pos),
		Name:   s.Name,
		Module: s.Module,
	}
}
