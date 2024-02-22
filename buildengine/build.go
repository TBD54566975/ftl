package buildengine

import (
	"context"

	"github.com/TBD54566975/ftl/backend/schema"
)

// A Module is a ModuleConfig with its dependencies populated.
type Module struct {
	ModuleConfig
	Dependencies []string
}

// Build a module in the given directory given the schema and module config.
func Build(ctx context.Context, schema *schema.Schema, config Module) error {
	panic("not implemented")
}
