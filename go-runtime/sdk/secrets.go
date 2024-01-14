package sdk

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// SecretType is a type that can be used as a secret value.
//
// Supported types are currently limited, but will eventually be extended to
// allow any type that FTL supports, including structs.
type SecretType interface {
	string | int | float64 | bool |
		[]string | []int | []float64 | []bool | []byte
	map[string]string | map[string]int | map[string]float64 | map[string]bool | map[string][]byte
}

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
func (c *SecretValue[Type]) Get() (out Type) {
	value, ok := os.LookupEnv(fmt.Sprintf("FTL_SECRET_%s_%s", strings.ToUpper(c.module), strings.ToUpper(c.name)))
	if !ok {
		return out
	}
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		panic(fmt.Errorf("failed to parse %s: %w", c, err))
	}
	return
}
