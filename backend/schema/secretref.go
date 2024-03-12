package schema

import (
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

// SecretRef is a reference to a Secret.
type SecretRef = AbstractRef[schemapb.SecretRef]

var _ Type = (*SecretRef)(nil)

func ParseSecretRef(ref string) (*SecretRef, error) { return ParseRef[schemapb.SecretRef](ref) }

func SecretRefFromProto(s *schemapb.SecretRef) *SecretRef {
	return &SecretRef{
		Pos:    posFromProto(s.Pos),
		Name:   s.Name,
		Module: s.Module,
	}
}

func secretRefListToSchema(s []*schemapb.SecretRef) []*SecretRef {
	var out []*SecretRef
	for _, n := range s {
		out = append(out, SecretRefFromProto(n))
	}
	return out
}
