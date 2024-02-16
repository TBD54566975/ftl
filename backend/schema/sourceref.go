package schema

import (
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

// SourceRef is a reference to a Source.
type SourceRef = AbstractRef[schemapb.SourceRef]

var _ Type = (*SourceRef)(nil)

func ParseSourceRef(ref string) (*SourceRef, error) { return ParseRef[schemapb.SourceRef](ref) }

func SourceRefFromProto(s *schemapb.SourceRef) *SourceRef {
	return &SourceRef{
		Pos:    posFromProto(s.Pos),
		Name:   s.Name,
		Module: s.Module,
	}
}

func sourceRefListToSchema(s []*schemapb.SourceRef) []*SourceRef {
	var out []*SourceRef
	for _, n := range s {
		out = append(out, SourceRefFromProto(n))
	}
	return out
}
