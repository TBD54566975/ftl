//nolint:revive
package model

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
)

func NewRunnerKey(hostname string, port string) RunnerKey {
	return keyType[runnerKey]{
		Hostname: hostname,
		Port:     port,
		Suffix:   rand.Intn(10000),
	}
}
func NewLocalRunnerKey(suffix int) RunnerKey {
	return keyType[runnerKey]{
		Suffix: suffix,
	}
}
func ParseRunnerKey(key string) (RunnerKey, error)   { return parseKey[RunnerKey](key, true) }
func ParseRunnerDBKey(key string) (RunnerKey, error) { return parseKey[RunnerKey](key, false) }

type runnerKey struct{}
type RunnerKey = keyType[runnerKey]

func NewControllerKey(hostname string, port string) ControllerKey {
	return keyType[controllerKey]{
		Hostname: hostname,
		Port:     port,
		Suffix:   rand.Intn(10000),
	}
}
func ParseControllerKey(key string) (ControllerKey, error) { return parseKey[ControllerKey](key, true) }
func ParseControllerDBKey(key string) (ControllerKey, error) {
	return parseKey[ControllerKey](key, false)
}

type controllerKey struct{}
type ControllerKey = keyType[controllerKey]

func parseKey[KT keyType[U], U any](key string, includesKind bool) (KT, error) {
	// Expected style: [<kind>-]<host>-<port>-<suffix> or [<kind>-]<suffix>

	components := strings.Split(key, "-")
	if includesKind {
		//
		if len(components) == 0 {
			return KT{}, fmt.Errorf("expected a prefix for key: %s", key)
		}
		kind := kindFromType[U]()
		if components[0] != kind {
			return KT{}, fmt.Errorf("unexpected prefix for key: %s", key)
		}
		components = components[1:]
	}

	switch {
	case len(components) == 1:
		//style: [<kind>-]<suffix>

		suffix, err := strconv.Atoi(components[len(components)-1])
		if err != nil {
			return KT{}, fmt.Errorf("invalid suffix for key: %s", key)
		}

		return KT{
			Suffix: suffix,
		}, nil
	case len(components) >= 3:
		//style: [<kind>-]<host>-<port>-<suffix>

		suffix, err := strconv.Atoi(components[len(components)-1])
		if err != nil {
			return KT{}, fmt.Errorf("invalid suffix for key: %s", key)
		}

		port := components[len(components)-2]
		host := strings.Join(components[:len(components)-2], "-")

		return KT{
			Hostname: host,
			Port:     port,
			Suffix:   suffix,
		}, nil
	default:
		return KT{}, fmt.Errorf("expected more components in key: %s", key)
	}

}

// Helper type to avoid having to write a bunch of boilerplate. It relies on T being a
// named struct in the form <name>Key, eg. "runnerKey"
type keyType[T any] struct {
	Hostname string
	Port     string
	Suffix   int
}

func (d keyType[T]) Value() (driver.Value, error) {
	return d.string(false), nil
}

var _ sql.Scanner = (*keyType[int])(nil)
var _ driver.Valuer = (*keyType[int])(nil)

// Scan from DB representation.
func (d *keyType[T]) Scan(src any) error {
	input, ok := src.(string)
	if !ok {
		return fmt.Errorf("expected key to be a string but it's a %T", src)
	}
	key, err := parseKey[keyType[T]](input, false)
	if err != nil {
		return err
	}
	*d = key
	return nil
}

func (d keyType[T]) Kind() string { return kindFromType[T]() }

func (d keyType[T]) String() string {
	return d.string(true)
}

func (d keyType[T]) string(includeKind bool) string {
	var prefix string
	if includeKind {
		prefix = fmt.Sprintf("%s-", d.Kind())
	}
	if d.Hostname == "" {
		return fmt.Sprintf("%s%04d", prefix, d.Suffix)
	}
	return fmt.Sprintf("%s%s-%s-%04d", prefix, d.Hostname, d.Port, d.Suffix)
}

func (d keyType[T]) MarshalText() ([]byte, error) { return []byte(d.String()), nil }
func (d *keyType[T]) UnmarshalText(bytes []byte) error {
	fmt.Printf("marshal text: %s\n", string(bytes))
	id, err := parseKey[keyType[T]](string(bytes), true)
	if err != nil {
		return err
	}
	*d = id
	return nil
}

func kindFromType[T any]() string {
	var zero T
	return strings.ToLower(strings.TrimSuffix(reflect.TypeOf(zero).Name(), "Key")[:1])
}
