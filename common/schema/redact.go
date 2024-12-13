package schema

import "github.com/TBD54566975/ftl/common/reflect"

type Redactable interface {
	Redact()
}

// Redact clones n and recursively removes sensitive information from it.
//
// Any node that implements the Redactable interface will have its Redact method called.
func Redact[T Node](n T) T {
	n = reflect.DeepCopy(n)
	_ = Visit(n, func(n Node, next func() error) error { //nolint:errcheck
		if redactable, ok := n.(Redactable); ok {
			redactable.Redact()
		}
		return next()
	})
	return n
}