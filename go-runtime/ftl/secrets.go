package ftl

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/common/configuration"
)

// SecretType is a type that can be used as a secret value.
type SecretType interface{ any }

// Secret declares a typed secret for the current module.
func Secret[Type SecretType](name string) SecretValue[Type] {
	module := callerModule()
	return SecretValue[Type]{module, name}
}

// SecretValue is a typed secret for the current module.
type SecretValue[Type SecretType] struct {
	module string
	name   string
}

func (s *SecretValue[Type]) String() string {
	return fmt.Sprintf("secret %s.%s", s.module, s.name)
}

// Get returns the value of the secret from FTL.
func (s *SecretValue[Type]) Get(ctx context.Context) (out Type) {
	sm := configuration.SecretsFromContext(ctx)
	if err := sm.Get(ctx, configuration.NewRef(s.module, s.name), &out); err != nil {
		panic(fmt.Errorf("failed to get %s: %w", s, err))
	}
	return
}
