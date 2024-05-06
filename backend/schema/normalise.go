package schema

import (
	"github.com/TBD54566975/ftl/internal/reflect"
)

// Normalise clones and normalises (zeroes) positional information in schema Nodes.
func Normalise[T Node](n T) T {
	ni := reflect.DeepCopy(n)
	_ = Visit(ni, func(n Node, next func() error) error {
		switch n := n.(type) {
		case *Bool:
			n.Bool = false

		case *Float:
			n.Float = false

		case *Int:
			n.Int = false

		case *String:
			n.Str = false

		case *Any:
			n.Any = false

		case *Unit:
			n.Unit = true

		case *Time:
			n.Time = false

		default: // Normally we don't default for sum types, but this is just for tests and will be immediately obvious.
		}
		return next()
	})
	return ni //nolint:forcetypeassert
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
