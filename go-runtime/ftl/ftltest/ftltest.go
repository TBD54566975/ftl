// Package ftltest contains test utilities for the ftl package.
package ftltest

import (
	"context"

	"github.com/alecthomas/kong"

	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/internal/log"
)

// Context suitable for use in testing FTL verbs.
func Context() context.Context {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	cr := &cf.ProjectConfigResolver[cf.Configuration]{Config: []string{}}
	_ = kong.ApplyDefaults(cr)
	cm, err := cf.NewConfigurationManager(ctx, cr)
	if err != nil {
		panic(err)
	}
	ctx = cf.ContextWithConfig(ctx, cm)
	sr := &cf.ProjectConfigResolver[cf.Secrets]{Config: []string{}}
	_ = kong.ApplyDefaults(sr)
	sm, err := cf.NewSecretsManager(ctx, sr)
	if err != nil {
		panic(err)
	}
	ctx = cf.ContextWithSecrets(ctx, sm)
	return ctx
}
