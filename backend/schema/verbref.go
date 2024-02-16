package schema

import (
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

// VerbRef is a reference to a Verb.
type VerbRef = AbstractRef[schemapb.VerbRef]

var _ Type = (*VerbRef)(nil)

func ParseVerbRef(ref string) (*VerbRef, error) { return ParseRef[schemapb.VerbRef](ref) }

func VerbRefFromProto(s *schemapb.VerbRef) *VerbRef {
	return &VerbRef{
		Pos:    posFromProto(s.Pos),
		Name:   s.Name,
		Module: s.Module,
	}
}

func verbRefListToSchema(s []*schemapb.VerbRef) []*VerbRef {
	var out []*VerbRef
	for _, n := range s {
		out = append(out, VerbRefFromProto(n))
	}
	return out
}
