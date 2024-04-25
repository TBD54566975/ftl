// Package ftltest contains test utilities for the ftl package.
package ftltest

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/modulecontext"
	"github.com/TBD54566975/ftl/internal/log"
)

// Context suitable for use in testing FTL verbs.
func Context() context.Context {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	context, err := modulecontext.FromEnvironment(ctx, ftl.Module())
	if err != nil {
		panic(err)
	}
	return context.ApplyToContext(ctx)
}
