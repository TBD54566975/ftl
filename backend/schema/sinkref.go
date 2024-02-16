package schema

import (
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

// SinkRef is a reference to a Sink.
type SinkRef = AbstractRef[schemapb.SinkRef]

var _ Type = (*SinkRef)(nil)

func ParseSinkRef(ref string) (*SinkRef, error) { return ParseRef[schemapb.SinkRef](ref) }

func SinkRefFromProto(s *schemapb.SinkRef) *SinkRef {
	return &SinkRef{
		Pos:    posFromProto(s.Pos),
		Name:   s.Name,
		Module: s.Module,
	}
}

func sinkRefListToSchema(s []*schemapb.SinkRef) []*SinkRef {
	var out []*SinkRef
	for _, n := range s {
		out = append(out, SinkRefFromProto(n))
	}
	return out
}
