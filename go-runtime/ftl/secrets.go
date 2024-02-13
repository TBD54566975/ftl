package ftl

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// SecretType is a type that can be used as a secret value.
type SecretType interface{ any }

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
func (s *SecretValue[Type]) Get() (out Type) {
	value, ok := os.LookupEnv(fmt.Sprintf("FTL_SECRET_%s_%s", strings.ToUpper(s.module), strings.ToUpper(s.name)))
	if !ok {
		return out
	}
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		panic(fmt.Errorf("failed to parse %s: %w", s, err))
	}
	return
}
