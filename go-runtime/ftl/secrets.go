package ftl

import (
	"context"
	"fmt"

	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/go-runtime/internal"
)

// SecretType is a type that can be used as a secret value.
type SecretType interface{ any }

// Secret declares a typed secret for the current module.
type Secret[T SecretType] struct {
	reflection.Ref
}

func (s Secret[T]) String() string { return fmt.Sprintf("secret \"%s\"", s.Ref) }

func (s Secret[T]) GoString() string {
	var t T
	return fmt.Sprintf("ftl.Secret[%T](\"%s\")", t, s.Ref)
}

// Get returns the value of the secret from FTL.
func (s Secret[T]) Get(ctx context.Context) (out T) {
	if err := internal.FromContext(ctx).GetSecret(ctx, s.Name, &out); err != nil {
		panic(fmt.Errorf("failed to get %s: %w", s, err))
	}
	return
}
