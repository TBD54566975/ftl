package schema

import "golang.design/x/reflect"

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

	case *DataRef:
		c.Pos = zero

	case *Field:
		c.Pos = zero
		c.Type = Normalise(c.Type)

	case *Float:
		c.Float = false
		c.Pos = zero

	case *Int:
		c.Int = false
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

	case *Bytes:
		c.Bytes = false
		c.Pos = zero

	case *Verb:
		c.Pos = zero
		c.Request = Normalise(c.Request)
		c.Response = Normalise(c.Response)
		c.Metadata = normaliseSlice(c.Metadata)

	case *VerbRef:
		c.Pos = zero

	case *MetadataCalls:
		c.Pos = zero
		c.Calls = normaliseSlice(c.Calls)

	case *MetadataDatabases:
		c.Pos = zero
		c.Calls = normaliseSlice(c.Calls)

	case *MetadataIngress:
		c.Pos = zero
		c.Path = normaliseSlice(c.Path)

	case *Optional:
		c.Type = Normalise(c.Type)

	case *IngressPathLiteral:
		c.Pos = zero

	case *IngressPathParameter:
		c.Pos = zero

	case *SourceRef:
		c.Pos = zero

	case *SinkRef:
		c.Pos = zero

	case Decl, Metadata, IngressPathComponent, Type: // Can never occur in reality, but here to satisfy the sum-type check.
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
