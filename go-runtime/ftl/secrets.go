package ftl

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/go-runtime/internal"
)

// SecretType is a type that can be used as a secret value.
type SecretType interface{ any }

// Secret declares a typed secret for the current module.
func Secret[T SecretType](name string) SecretValue[T] {
	module := callerModule()
	return SecretValue[T]{reflection.Ref{Module: module, Name: name}}
}

// SecretValue is a typed secret for the current module.
type SecretValue[T SecretType] struct {
	reflection.Ref
}

func (s SecretValue[T]) String() string { return fmt.Sprintf("secret \"%s\"", s.Ref) }

func (s SecretValue[T]) GoString() string {
	var t T
	return fmt.Sprintf("ftl.SecretValue[%T](\"%s\")", t, s.Ref)
}

// Get returns the value of the secret from FTL.
func (s SecretValue[T]) Get(ctx context.Context) (out T) {
	if err := internal.FromContext(ctx).GetSecret(ctx, s.Name, &out); err != nil {
		panic(fmt.Errorf("failed to get %s: %w", s, err))
	}
	return
}
