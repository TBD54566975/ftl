package ftl

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/go-runtime/internal"
)

// ConfigType is a type that can be used as a configuration value.
type ConfigType interface{ any }

// Config is a typed configuration key for the current module.
type Config[T ConfigType] struct {
	reflection.Ref
}

func (c Config[T]) String() string { return fmt.Sprintf("config \"%s\"", c.Ref) }

func (c Config[T]) GoString() string {
	var t T
	return fmt.Sprintf("ftl.Config[%T](\"%s\")", t, c.Ref)
}

// Get returns the value of the configuration key from FTL.
func (c Config[T]) Get(ctx context.Context) (out T) {
	err := internal.FromContext(ctx).GetConfig(ctx, c.Name, &out)
	if err != nil {
		panic(fmt.Errorf("failed to get %s: %w", c, err))
	}
	return
}
