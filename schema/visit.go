package schema

import (
	"github.com/alecthomas/errors"
	"github.com/alecthomas/participle/v2/lexer"
)

// Visit all nodes in the schema.
func Visit(n Node, visit func(n Node, next func() error) error) error {
	return visit(n, func() error {
		for _, child := range n.schemaChildren() {
			if err := Visit(child, visit); err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	})
}

// Normalise a Node.
func Normalise[T Node](n T) T {
	var zero lexer.Position
	var ni Node = n
	switch c := ni.(type) {
	case Schema:
		c.Pos = zero
		c.Modules = normaliseSlice(c.Modules)
		ni = c
	case Module:
		c.Pos = zero
		c.Decls = normaliseSlice(c.Decls)
		ni = c
	case Array:
		c.Pos = zero
		c.Element = Normalise(c.Element)
		ni = c
	case Bool:
		c.Pos = zero
		ni = c
	case Data:
		c.Pos = zero
		c.Fields = normaliseSlice(c.Fields)
		c.Metadata = normaliseSlice(c.Metadata)
		ni = c
	case DataRef:
		c.Pos = zero
		ni = c
	case Field:
		c.Pos = zero
		c.Type = Normalise(c.Type)
		ni = c
	case Float:
		c.Pos = zero
		ni = c
	case Int:
		c.Pos = zero
		ni = c
	case Map:
		c.Pos = zero
		c.Key = Normalise(c.Key)
		c.Value = Normalise(c.Value)
		ni = c
	case String:
		c.Pos = zero
		ni = c
	case Verb:
		c.Pos = zero
		c.Request = Normalise(c.Request)
		c.Response = Normalise(c.Response)
		c.Metadata = normaliseSlice(c.Metadata)
		ni = c
	case VerbRef:
		c.Pos = zero
		ni = c
	case MetadataCalls:
		c.Pos = zero
		c.Calls = normaliseSlice(c.Calls)
		ni = c
	case Decl, Metadata, Type: // Can never occur in reality, but here to satisfy the sum-type check.
		panic("??")
	}
	if ni == nil {
		panic("ni is nil, it must be set in the switch statement")
	}
	return ni.(T) //nolint:forcetypeassert
}

func normaliseSlice[T Node](in []T) []T {
	if in == nil {
		return nil
	}
	out := make([]T, len(in))
	for i, n := range in {
		out[i] = Normalise(n)
	}
	return out
}
