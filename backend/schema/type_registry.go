package schema

import (
	"context"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type contextKeyTypeRegistry struct{}

// ContextWithTypeRegistry adds a type registry to the given context.
func ContextWithTypeRegistry(ctx context.Context, r *schemapb.TypeRegistry) context.Context {
	return context.WithValue(ctx, contextKeyTypeRegistry{}, r)
}

// TypeRegistry is a registry of types that can be resolved to a schema type at runtime.
// It also records sum types and their variants, for use in encoding and decoding.
type TypeRegistry struct {
	// SchemaTypes associates a type name with a schema type.
	SchemaTypes map[string]Type `protobuf:"1"`
	// SumTypes associates a sum type discriminator type name with its variant type names.
	SumTypes map[string]SumTypeVariants `protobuf:"2"`
}

// NewTypeRegistry creates a new type registry.
// The type registry is used to instantiate types by their qualified name at runtime.
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		SchemaTypes: make(map[string]Type),
		SumTypes:    make(map[string]SumTypeVariants),
	}
}

// RegisterSumType registers a Go sum type with the type registry. Sum types are represented as enums in the
// FTL schema.
func (t *TypeRegistry) RegisterSumType(discriminator string, variants map[string]Type) {
	var values []string
	for name, vt := range variants {
		values = append(values, name)
		t.SchemaTypes[name] = vt
	}
	t.SumTypes[discriminator] = SumTypeVariants{Value: values}
}

func (t *TypeRegistry) ToProto() *schemapb.TypeRegistry {
	typespb := make(map[string]*schemapb.Type, len(t.SchemaTypes))
	for k, v := range t.SchemaTypes {
		typespb[k] = typeToProto(v)
	}

	return &schemapb.TypeRegistry{
		SumTypes:    sumTypeVariantsToProto(t.SumTypes),
		SchemaTypes: typespb,
	}
}

type SumTypeVariants struct {
	// Value is a list of variant names for the sum type.
	Value []string `protobuf:"1"`
}

func (s *SumTypeVariants) ToProto() *schemapb.SumTypeVariants {
	return &schemapb.SumTypeVariants{Value: s.Value}
}

func sumTypeVariantsToProto(v map[string]SumTypeVariants) map[string]*schemapb.SumTypeVariants {
	out := make(map[string]*schemapb.SumTypeVariants, len(v))
	for k, v := range v {
		out[k] = v.ToProto()
	}
	return out
}
