package configuration

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
)

type EnvarType interface{ prefix() string }

type EnvarTypeConfig struct{}

func (EnvarTypeConfig) prefix() string { return "FTL_CONFIG_" }

type EnvarTypeSecrets struct{}

func (EnvarTypeSecrets) prefix() string { return "FTL_SECRET_" }

// EnvarProvider is a configuration provider that reads from environment variables.
type EnvarProvider[T EnvarType] struct {
	Envar bool `help:"Print configuration as environment variables." xor:"configwriter" group:"Provider:"`
}

var _ MutableProvider = EnvarProvider[EnvarTypeConfig]{}

func (EnvarProvider[T]) Key() string { return "envar" }

func (e EnvarProvider[T]) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	// FTL_<type>_[<module>]_<name> where <module> and <name> are base64 encoded.
	envar := e.key(ref)

	value, ok := os.LookupEnv(envar)
	if ok {
		return base64.RawStdEncoding.DecodeString(value)
	}
	return nil, fmt.Errorf("environment variable %q is not set: %w", envar, ErrNotFound)
}

func (e EnvarProvider[T]) Delete(ctx context.Context, ref Ref) error {
	return nil
}

func (e EnvarProvider[T]) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	envar := e.key(ref)
	fmt.Printf("%s=%s\n", envar, base64.RawStdEncoding.EncodeToString(value))
	return &url.URL{Scheme: "envar", Host: ref.Name}, nil
}

func (e EnvarProvider[T]) Writer() bool { return e.Envar }

func (e EnvarProvider[T]) key(ref Ref) string {
	key := e.prefix()
	if m, ok := ref.Module.Get(); ok {
		key += base64.RawStdEncoding.EncodeToString([]byte(m)) + "_"
	}
	key += base64.RawStdEncoding.EncodeToString([]byte(ref.Name))
	return key
}

func (EnvarProvider[T]) prefix() string {
	var t T
	return t.prefix()
}

// I don't think there's a need to parse environment variables, but let's keep
// this around for a bit just in case, as it was a PITA to write.
//
// func (e EnvarProvider[T]) entryForEnvar(env string) (Entry, error) {
// 	parts := strings.SplitN(env, "=", 2)
// 	if !strings.HasPrefix(parts[0], e.prefix()) {
// 		return Entry{}, fmt.Errorf("invalid environment variable %q", parts[0])
// 	}
// 	accessor, err := url.Parse(parts[1])
// 	if err != nil {
// 		return Entry{}, fmt.Errorf("invalid URL %q: %w", parts[1], err)
// 	}
// 	// FTL_<type>_[<module>]_<name>
// 	nameParts := strings.SplitN(parts[0], "_", 4)
// 	if len(nameParts) < 4 {
// 		return Entry{}, fmt.Errorf("invalid environment variable %q", parts[0])
// 	}
// 	var module optional.Option[string]
// 	if nameParts[2] != "" {
// 		decoded, err := base64.RawStdEncoding.DecodeString(nameParts[2])
// 		if err != nil {
// 			return Entry{}, fmt.Errorf("invalid encoded module %q: %w", nameParts[2], err)
// 		}
// 		module = optional.Some(string(decoded))
// 	}
// 	decoded, err := base64.RawStdEncoding.DecodeString(nameParts[3])
// 	if err != nil {
// 		return Entry{}, fmt.Errorf("invalid encoded name %q: %w", nameParts[3], err)
// 	}
// 	return Entry{
// 		Ref:      Ref{module, string(decoded)},
// 		Accessor: accessor,
// 	}, nil
// }
