package sdk

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
)

// ConfigType is a type that can be used as a configuration value.
//
// Supported types are currently limited, but will eventually be extended to
// allow any type that FTL supports, including structs.
type ConfigType interface {
	string | int | float64 | bool |
		[]string | []int | []float64 | []bool | []byte |
		map[string]string | map[string]int | map[string]float64 | map[string]bool | map[string][]byte
}

// Config declares a typed configuration key for the current module.
func Config[T ConfigType](name string) ConfigValue[T] {
	module := callerModule()
	return ConfigValue[T]{module, name}
}

// ConfigValue is a typed configuration key for the current module.
type ConfigValue[T ConfigType] struct {
	module string
	name   string
}

func (c *ConfigValue[T]) String() string {
	return fmt.Sprintf("config %s.%s", c.module, c.name)
}

// Get returns the value of the configuration key from FTL.
func (c *ConfigValue[T]) Get() (out T) {
	value, ok := os.LookupEnv(fmt.Sprintf("FTL_CONFIG_%s_%s", strings.ToUpper(c.module), strings.ToUpper(c.name)))
	if !ok {
		return out
	}
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		panic(fmt.Errorf("failed to parse %s value %q: %w", c, value, err))
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
	if strings.HasPrefix(module, "github.com/TBD54566975/ftl/go-runtime/sdk") {
		return "testing"
	}
	if !strings.HasPrefix(module, "ftl/") {
		panic(fmt.Sprintf("must be called from an FTL module not %s", module))
	}
	return strings.Split(strings.Split(module, "/")[1], ".")[0]
}
