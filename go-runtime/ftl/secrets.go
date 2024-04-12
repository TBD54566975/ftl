package ftl

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/internal/log"
)

var globalSecretsManagerOnce sync.Once

// globalSecretsManager returns a global secrets manager instance.
func globalSecretsManager() (manager *cf.Manager[cf.Secrets]) {
	globalSecretsManagerOnce.Do(func() {
		var configs []string
		if envar, ok := os.LookupEnv("FTL_CONFIG"); ok {
			configs = strings.Split(envar, ",")
		}
		config := cf.DefaultSecretsMixin{}
		var err error
		ctx := log.ContextWithNewDefaultLogger(context.Background())
		manager, err = config.NewSecretsManager(ctx, cf.ProjectConfigResolver[cf.Secrets]{Config: configs})
		if err != nil {
			panic("failed to create global secrets manager: " + err.Error())
		}
	})
	return manager
}

// Secret declares a typed secret for the current module.
func Secret[T any](name string) SecretValue[T] {
	module := callerModule()
	return SecretValue[T]{module, name}
}

// SecretValue is a typed secret for the current module.
type SecretValue[T any] struct {
	module string
	name   string
}

func (s SecretValue[T]) String() string { return fmt.Sprintf("secret \"%s.%s\"", s.module, s.name) }

func (s SecretValue[T]) GoString() string {
	var t T
	return fmt.Sprintf("ftl.SecretValue[%T](\"%s.%s\")", t, s.module, s.name)
}

// Get returns the value of the secret from FTL.
func (s SecretValue[T]) Get() (out T) {
	sm := globalSecretsManager()
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	if err := sm.Get(ctx, cf.NewRef(s.module, s.name), &out); err != nil {
		panic(fmt.Errorf("failed to get secrets %s.%s: %w", s.module, s.name, err))
	}
	return
}
