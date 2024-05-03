package modulecontext

import (
	"context"
	"fmt"
	"os"
	"strings"

	cf "github.com/TBD54566975/ftl/common/configuration"
)

// FromEnvironment creates a ModuleContext from the local environment.
//
// This is useful for testing and development, where the environment is used to provide
// configurations, secrets and DSNs. The context is built from a combination of
// the ftl-project.toml file and (for now) environment variables.
func FromEnvironment(ctx context.Context, module string, isTesting bool) (*ModuleContext, error) {
	// TODO: split this func into separate purposes: explicitly reading a particular project file, and reading DSNs from environment
	var moduleCtx *ModuleContext
	if isTesting {
		moduleCtx = NewForTesting()
	} else {
		moduleCtx = New()
	}

	cm, err := cf.NewDefaultConfigurationManagerFromEnvironment(ctx)
	if err != nil {
		return nil, err
	}
	configs, err := cm.MapForModule(ctx, module)
	if err != nil {
		return nil, err
	}
	for name, data := range configs {
		moduleCtx.SetConfigData(name, data)
	}

	sm, err := cf.NewDefaultSecretsManagerFromEnvironment(ctx)
	if err != nil {
		return nil, err
	}
	secrets, err := sm.MapForModule(ctx, module)
	if err != nil {
		return nil, err
	}
	for name, data := range secrets {
		moduleCtx.SetSecretData(name, data)
	}

	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "FTL_POSTGRES_DSN_") {
			continue
		}
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid DSN environment variable: %s", entry)
		}
		key := parts[0]
		value := parts[1]
		// FTL_POSTGRES_DSN_MODULE_DBNAME
		parts = strings.Split(key, "_")
		if len(parts) != 5 {
			return nil, fmt.Errorf("invalid DSN environment variable: %s", entry)
		}
		moduleName := parts[3]
		dbName := parts[4]
		if !strings.EqualFold(moduleName, module) {
			continue
		}
		if err := moduleCtx.AddDatabase(strings.ToLower(dbName), DBTypePostgres, value); err != nil {
			return nil, err
		}
	}
	return moduleCtx, nil
}
