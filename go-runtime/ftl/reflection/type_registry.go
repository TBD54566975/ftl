package reflection

import (
	"reflect"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/alecthomas/types/optional"
)

// TypeRegistry is used for dynamic type resolution at runtime. It stores associations between sum type discriminators
// and their variants, for use in encoding and decoding.
//
// FTL manages the type registry for you, so you don't need to create one yourself.
type TypeRegistry struct {
	sumTypes                 map[reflect.Type][]sumTypeVariant
	variantsToDiscriminators map[reflect.Type]reflect.Type
	fsm                      map[string]ReflectedFSM
	externalTypes            map[reflect.Type]struct{}
}

type sumTypeVariant struct {
	name   string
	goType reflect.Type
}

// Registree is a function that registers types with a [TypeRegistry].
type Registree func(t *TypeRegistry)

// SumType adds a sum type and its variants to the type registry.
func SumType[Discriminator any](variants ...Discriminator) Registree {
	return func(t *TypeRegistry) {
		variantMap := map[string]reflect.Type{}
		for _, v := range variants {
			ref := TypeRefFromValue(v)
			variantMap[ref.Name] = reflect.TypeOf(v)
		}
		t.registerSumType(reflect.TypeFor[Discriminator](), variantMap)
	}
}

// Transition represents a transition between two states in an FSM.
type Transition struct {
	From reflect.Value
	To   reflect.Value
}

type ReflectedFSM struct {
	Transitions []Transition
	Schema      *schema.FSM
}

// FSM adds a finite state machine to the type registry.
func FSM(name string, transitions ...Transition) Registree {
	return func(t *TypeRegistry) { t.registerFSM(name, transitions) }
}

// ExternalType adds a non-FTL type to the type registry.
func ExternalType(goType any) Registree {
	return func(t *TypeRegistry) {
		typ := reflect.TypeOf(goType)
		t.externalTypes[typ] = struct{}{}
	}
}

// newTypeRegistry creates a new [TypeRegistry] for instantiating types by their qualified
// name at runtime.
func newTypeRegistry(options ...Registree) *TypeRegistry {
	t := &TypeRegistry{
		sumTypes:                 map[reflect.Type][]sumTypeVariant{},
		variantsToDiscriminators: map[reflect.Type]reflect.Type{},
		fsm:                      map[string]ReflectedFSM{},
		externalTypes:            map[reflect.Type]struct{}{},
	}
	for _, o := range options {
		o(t)
	}
	return t
}

// registerSumType registers a Go sum type with the type registry.
//
// Sum types are represented as enums in the FTL schema.
func (t *TypeRegistry) registerSumType(discriminator reflect.Type, variants map[string]reflect.Type) {
	var values []sumTypeVariant
	for name, v := range variants {
		t.variantsToDiscriminators[v] = discriminator
		values = append(values, sumTypeVariant{
			name:   name,
			goType: v,
		})
	}
	t.sumTypes[discriminator] = values
}

func (t *TypeRegistry) registerFSM(name string, transitions []Transition) {
	fsm := &schema.FSM{Name: name}
	for _, transition := range transitions {
		if !transition.From.IsValid() {
			fsm.Start = append(fsm.Start, FuncRef(transition.To.Interface()).ToSchema())
		} else {
			fsm.Transitions = append(fsm.Transitions, &schema.FSMTransition{
				From: FuncRef(transition.From.Interface()).ToSchema(),
				To:   FuncRef(transition.To.Interface()).ToSchema(),
			})
		}
	}
	t.fsm[name] = ReflectedFSM{
		Transitions: transitions,
		Schema:      fsm,
	}
}

func (t *TypeRegistry) isSumTypeDiscriminator(discriminator reflect.Type) bool {
	return t.getSumTypeVariants(discriminator).Ok()
}

func (t *TypeRegistry) getFSM(name string) optional.Option[ReflectedFSM] {
	return optional.Zero(t.fsm[name])
}

func (t *TypeRegistry) getDiscriminatorByVariant(variant reflect.Type) optional.Option[reflect.Type] {
	return optional.Zero(t.variantsToDiscriminators[variant])
}

func (t *TypeRegistry) getVariantByName(discriminator reflect.Type, name string) optional.Option[reflect.Type] {
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

func (t *TypeRegistry) getVariantByType(discriminator reflect.Type, variantType reflect.Type) optional.Option[string] {
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
	return optional.Zero(t.sumTypes[discriminator])
}
