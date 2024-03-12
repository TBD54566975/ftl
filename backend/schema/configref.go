package schema

import (
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

// ConfigRef is a reference to a Config.
type ConfigRef = AbstractRef[schemapb.ConfigRef]

var _ Type = (*ConfigRef)(nil)

func ParseConfigRef(ref string) (*ConfigRef, error) { return ParseRef[schemapb.ConfigRef](ref) }

func ConfigRefFromProto(s *schemapb.ConfigRef) *ConfigRef {
	return &ConfigRef{
		Pos:    posFromProto(s.Pos),
		Name:   s.Name,
		Module: s.Module,
	}
}

func configRefListToSchema(s []*schemapb.ConfigRef) []*ConfigRef {
	var out []*ConfigRef
	for _, n := range s {
		out = append(out, ConfigRefFromProto(n))
	}
	return out
}
