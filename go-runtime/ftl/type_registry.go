package ftl

import (
	"fmt"
	"reflect"
)

// TypeRegistry is a registry of types that can be instantiated by their qualified name.
// It also records sum types and their variants, for use in encoding and decoding.
//
// FTL manages the type registry for you, so you don't need to create one yourself.
type TypeRegistry struct {
	// sumTypes associates a sum type discriminator with its variants
	sumTypes map[string][]sumTypeVariant
	types    map[string]reflect.Type
}

type sumTypeVariant struct {
	name     string
	typeName string
}

// NewTypeRegistry creates a new type registry.
// The type registry is used to instantiate types by their qualified name at runtime.
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{types: map[string]reflect.Type{}, sumTypes: map[string][]sumTypeVariant{}}
}

// New creates a new instance of the type from the qualified type name.
func (t *TypeRegistry) New(name string) (any, error) {
	typ, ok := t.types[name]
	if !ok {
		return nil, fmt.Errorf("type %q not registered", name)
	}
	return reflect.New(typ).Interface(), nil
}

// RegisterSumType registers a Go sum type with the type registry. Sum types are represented as enums in the
// FTL schema.
func (t *TypeRegistry) RegisterSumType(discriminator reflect.Type, variants map[string]reflect.Type) {
	dFqName := discriminator.PkgPath() + "." + discriminator.Name()
	t.types[dFqName] = discriminator
	t.sumTypes[dFqName] = make([]sumTypeVariant, 0, len(variants))

	for name, v := range variants {
		vFqName := v.PkgPath() + "." + v.Name()
		t.types[vFqName] = v
		t.sumTypes[dFqName] = append(t.sumTypes[dFqName], sumTypeVariant{name: name, typeName: vFqName})
	}
}
