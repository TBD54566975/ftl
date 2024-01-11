package schema

import (
	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

// DataRef is a reference to a data structure.
type DataRef = AbstractRef[schemapb.DataRef]

var _ Type = (*DataRef)(nil)

func ParseDataRef(ref string) (*DataRef, error) { return ParseRef[schemapb.DataRef](ref) }

func DataRefFromProto(s *schemapb.DataRef) *DataRef {
	return &DataRef{
		Pos:    posFromProto(s.Pos),
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
