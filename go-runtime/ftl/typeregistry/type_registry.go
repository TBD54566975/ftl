package typeregistry

import (
	"context"
	"fmt"
	"reflect"
)

type contextKeyTypeRegistry struct{}

// ContextWithTypeRegistry adds a type registry to the given context.
func ContextWithTypeRegistry(ctx context.Context, r *TypeRegistry) context.Context {
	return context.WithValue(ctx, contextKeyTypeRegistry{}, r)
}

// TypeRegistry is a registry of types that can be instantiated by their qualified name.
// It also records sum types and their variants, for use in encoding and decoding.
//
// FTL manages the type registry for you, so you don't need to create one yourself
type TypeRegistry struct {
	// GoTypes associates a type name with a Go type.
	GoTypes map[string]reflect.Type
	// SumTypes associates a sum type discriminator type name with its variant type names.
	SumTypes map[string][]string
}

// NewTypeRegistry creates a new type registry.
// The type registry is used to instantiate types by their qualified name at runtime.
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		GoTypes:  make(map[string]reflect.Type),
		SumTypes: make(map[string][]string),
	}
}

// New creates a new instance of the type from the qualified type name.
func (t *TypeRegistry) New(name string) (any, error) {
	typ, ok := t.GoTypes[name]
	if !ok {
		return nil, fmt.Errorf("type %q not registered", name)
	}
	return reflect.New(typ).Interface(), nil
}

// RegisterSumType registers a Go sum type with the type registry. Sum types are represented as enums in the
// FTL schema.
func (t *TypeRegistry) RegisterSumType(discriminator reflect.Type, variants map[string]reflect.Type) {
	dFqName := discriminator.PkgPath() + "." + discriminator.Name()
	t.GoTypes[dFqName] = discriminator

	var values []string
	for name, v := range variants {
		values = append(values, name)
		vFqName := v.PkgPath() + "." + v.Name()
		t.GoTypes[vFqName] = v
	}
	t.SumTypes[dFqName] = values
}
