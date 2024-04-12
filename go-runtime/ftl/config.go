package ftl

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/internal/log"
)

var globalConfigurationManagerOnce sync.Once

// globalConfigurationManager returns a global configuration manager instance.
func globalConfigurationManager() (manager *cf.Manager[cf.Configuration]) {
	globalConfigurationManagerOnce.Do(func() {
		var configs []string
		if envar, ok := os.LookupEnv("FTL_CONFIG"); ok {
			configs = strings.Split(envar, ",")
		}
		config := cf.DefaultConfigMixin{}
		var err error
		ctx := log.ContextWithNewDefaultLogger(context.Background())
		manager, err = config.NewConfigurationManager(ctx, cf.ProjectConfigResolver[cf.Configuration]{Config: configs})
		if err != nil {
			panic("failed to create global configuration manager: " + err.Error())
		}
	})
	return manager
}

// Config loads a typed configuration value for the current module.
func Config[T any](name string) ConfigValue[T] {
	module := callerModule()
	return ConfigValue[T]{module, name}
}

// ConfigValue is a typed configuration key for the current module.
type ConfigValue[T any] struct {
	module string
	name   string
}

func (c ConfigValue[T]) String() string { return fmt.Sprintf("config \"%s.%s\"", c.module, c.name) }

func (c ConfigValue[T]) GoString() string {
	var t T
	return fmt.Sprintf("ftl.ConfigValue[%T](\"%s.%s\")", t, c.module, c.name)
}

// Get returns the value of the configuration key from FTL.
func (c ConfigValue[T]) Get() (out T) {
	cm := globalConfigurationManager()
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	if err := cm.Get(ctx, cf.NewRef(c.module, c.name), &out); err != nil {
		panic(fmt.Errorf("failed to get configuration %s.%s: %w", c.module, c.name, err))
	}
	return
}

func callerModule() string {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		panic("failed to get caller")
	}
	details := runtime.FuncForPC(pc)
	if details == nil {
		panic("failed to get caller")
	}
	module := details.Name()
	if strings.HasPrefix(module, "github.com/TBD54566975/ftl/go-runtime/ftl") {
		return "testing"
	}
	if !strings.HasPrefix(module, "ftl/") {
		panic("must be called from an FTL module not " + module)
	}
	return strings.Split(strings.Split(module, "/")[1], ".")[0]
}
