package schema

import "github.com/TBD54566975/ftl/internal/reflect"

// Normalise clones and normalises (zeroes) positional information in schema Nodes.
func Normalise[T Node](n T) T {
	n = reflect.DeepCopy(n)
	var zero Position
	var ni Node = n
	switch c := ni.(type) {
	case *Any:
		c.Any = false
		c.Pos = zero

	case *TypeParameter:
		c.Pos = zero

	case *Unit:
		c.Unit = true
		c.Pos = zero

	case *Schema:
		c.Pos = zero
		c.Modules = normaliseSlice(c.Modules)

	case *Module:
		c.Pos = zero
		c.Decls = normaliseSlice(c.Decls)

	case *Array:
		c.Pos = zero
		c.Element = Normalise(c.Element)

	case *Bool:
		c.Bool = false
		c.Pos = zero

	case *Data:
		c.Pos = zero
		c.TypeParameters = normaliseSlice(c.TypeParameters)
		c.Fields = normaliseSlice(c.Fields)
		c.Metadata = normaliseSlice(c.Metadata)

	case *Database:
		c.Pos = zero

	case *Ref:
		c.TypeParameters = normaliseSlice(c.TypeParameters)
		c.Pos = zero

	case *Enum:
		c.Pos = zero
		if c.Type != nil {
			c.Type = Normalise(c.Type)
		}
		c.Variants = normaliseSlice(c.Variants)

	case *EnumVariant:
		c.Pos = zero
		c.Value = Normalise(c.Value)

	case *TypeValue:
		c.Pos = zero
		c.Value = Normalise(c.Value)

	case *Field:
		c.Pos = zero
		c.Type = Normalise(c.Type)
		c.Metadata = normaliseSlice(c.Metadata)

	case *Float:
		c.Float = false
		c.Pos = zero

	case *Int:
		c.Int = false
		c.Pos = zero

	case *IntValue:
		c.Pos = zero

	case *Time:
		c.Time = false
		c.Pos = zero

	case *Map:
		c.Pos = zero
		c.Key = Normalise(c.Key)
		c.Value = Normalise(c.Value)

	case *String:
		c.Str = false
		c.Pos = zero

	case *StringValue:
		c.Pos = zero

	case *Bytes:
		c.Bytes = false
		c.Pos = zero

	case *Verb:
		c.Pos = zero
		c.Request = Normalise(c.Request)
		c.Response = Normalise(c.Response)
		c.Metadata = normaliseSlice(c.Metadata)

	case *MetadataCalls:
		c.Pos = zero
		c.Calls = normaliseSlice(c.Calls)

	case *MetadataDatabases:
		c.Pos = zero
		c.Calls = normaliseSlice(c.Calls)

	case *MetadataIngress:
		c.Pos = zero
		c.Path = normaliseSlice(c.Path)

	case *MetadataAlias:
		c.Pos = zero

	case *Optional:
		c.Type = Normalise(c.Type)

	case *IngressPathLiteral:
		c.Pos = zero

	case *IngressPathParameter:
		c.Pos = zero

	case *MetadataCronJob:
		c.Pos = zero

	case *Config:
		c.Pos = zero
		c.Type = Normalise(c.Type)

	case *Secret:
		c.Pos = zero
		c.Type = Normalise(c.Type)

	case Named, Symbol, Decl, Metadata, IngressPathComponent, Type, Value: // Can never occur in reality, but here to satisfy the sum-type check.
		panic("??")
	}
	return ni.(T) //nolint:forcetypeassert
}

func normaliseSlice[T Node](in []T) []T {
	if in == nil {
		return nil
	}
	var out []T
	for _, n := range in {
		out = append(out, Normalise(n))
	}
	return out
}
