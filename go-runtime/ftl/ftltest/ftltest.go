// Package ftltest contains test utilities for the ftl package.
package ftltest

import (
	"context"

	"github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/internal/log"
)

// Context suitable for use in testing FTL verbs.
func Context() context.Context {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	cm, err := configuration.DefaultConfigMixin{}.NewConfigurationManager(ctx)
	if err != nil {
		panic(err)
	}
	ctx = configuration.ContextWithConfig(ctx, cm)
	sm, err := configuration.DefaultSecretsMixin{}.NewSecretsManager(ctx)
	if err != nil {
		panic(err)
	}
	ctx = configuration.ContextWithSecrets(ctx, sm)
	return ctx
}
