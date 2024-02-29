package schema

import schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"

// EnumRef is a reference to an Enum.
type EnumRef = AbstractRef[schemapb.EnumRef]

var _ Type = (*EnumRef)(nil)

func ParseEnumRef(ref string) (*EnumRef, error) { return ParseRef[schemapb.EnumRef](ref) }

func EnumRefFromProto(s *schemapb.EnumRef) *EnumRef {
	return &EnumRef{
		Pos:    posFromProto(s.Pos),
		Name:   s.Name,
		Module: s.Module,
	}
}

func enumRefListToSchema(s []*schemapb.EnumRef) []*EnumRef {
	var out []*EnumRef
	for _, n := range s {
		out = append(out, EnumRefFromProto(n))
	}
	return out
}
