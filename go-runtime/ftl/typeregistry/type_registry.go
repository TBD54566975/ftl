package typeregistry

import (
	"context"
	"reflect"

	"github.com/alecthomas/types/optional"
)

type contextKeyTypeRegistry struct{}

// ContextWithTypeRegistry adds a type registry to the given context.
func ContextWithTypeRegistry(ctx context.Context, r *TypeRegistry) context.Context {
	return context.WithValue(ctx, contextKeyTypeRegistry{}, r)
}

// FromContext retrieves the secrets schema.TypeRegistry previously
// added to the context with [ContextWithTypeRegistry].
func FromContext(ctx context.Context) optional.Option[*TypeRegistry] {
	t, ok := ctx.Value(contextKeyTypeRegistry{}).(*TypeRegistry)
	if ok {
		return optional.Some(t)
	}
	return optional.None[*TypeRegistry]()
}

// TypeRegistry is used for dynamic type resolution at runtime. It stores associations between sum type discriminators
// and their variants, for use in encoding and decoding.
//
// FTL manages the type registry for you, so you don't need to create one yourself.
type TypeRegistry struct {
	sumTypes map[string][]sumTypeVariant
}

type sumTypeVariant struct {
	name   string
	goType reflect.Type
}

// NewTypeRegistry creates a new type registry.
// The type registry is used to instantiate types by their qualified name at runtime.
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		sumTypes: make(map[string][]sumTypeVariant),
	}
}

// RegisterSumType registers a Go sum type with the type registry. Sum types are represented as enums in the
// FTL schema.
func (t *TypeRegistry) RegisterSumType(discriminator reflect.Type, variants map[string]reflect.Type) {
	dFqName := discriminator.PkgPath() + "." + discriminator.Name()

	var values []sumTypeVariant
	for name, v := range variants {
		values = append(values, sumTypeVariant{
			name:   name,
			goType: v,
		})
	}
	t.sumTypes[dFqName] = values
}

func (t *TypeRegistry) IsSumTypeDiscriminator(discriminator reflect.Type) bool {
	return t.getSumTypeVariants(discriminator).Ok()
}

func (t *TypeRegistry) GetVariantByName(discriminator reflect.Type, name string) optional.Option[reflect.Type] {
	variants, ok := t.getSumTypeVariants(discriminator).Get()
	if !ok {
		return optional.None[reflect.Type]()
	}
	for _, v := range variants {
		if v.name == name {
			return optional.Some(v.goType)
		}
	}
	return optional.None[reflect.Type]()
}

func (t *TypeRegistry) GetVariantByType(discriminator reflect.Type, variantType reflect.Type) optional.Option[string] {
	variants, ok := t.getSumTypeVariants(discriminator).Get()
	if !ok {
		return optional.None[string]()
	}
	for _, v := range variants {
		if v.goType == variantType {
			return optional.Some(v.name)
		}
	}
	return optional.None[string]()
}

func (t *TypeRegistry) getSumTypeVariants(discriminator reflect.Type) optional.Option[[]sumTypeVariant] {
	dFqName := discriminator.PkgPath() + "." + discriminator.Name()
	variants, ok := t.sumTypes[dFqName]
	if !ok {
		return optional.None[[]sumTypeVariant]()
	}

	return optional.Some(variants)
}
