//nolint:revive
package model

import (
	"reflect"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/oklog/ulid/v2"
)

func NewDeploymentKey() DeploymentKey                      { return DeploymentKey(ulid.Make()) }
func ParseDeploymentKey(key string) (DeploymentKey, error) { return parseKey[DeploymentKey](key) }

type deploymentKey struct{}
type DeploymentKey = keyType[deploymentKey]

func NewRunnerKey() RunnerKey                      { return RunnerKey(ulid.Make()) }
func ParseRunnerKey(key string) (RunnerKey, error) { return parseKey[RunnerKey](key) }

type runnerKey struct{}
type RunnerKey = keyType[runnerKey]

func parseKey[KT keyType[U], U any](key string) (KT, error) {
	var zero KT
	kind := strings.TrimSuffix(reflect.TypeOf(*new(U)).Name(), "Key")
	prefix := "ftl:" + kind + ":"
	if !strings.HasPrefix(key, prefix) {
		return zero, errors.Errorf("invalid %s key %q", kind, key)
	}
	ulid, err := ulid.Parse(key[len(prefix):])
	if err != nil {
		return KT(ulid), err
	}
	return KT(ulid), nil
}

// Helper type to avoid having to write a bunch of boilerplate. It relies on T being a
// named struct in the form <name>Key, eg. "runnerKey"
type keyType[T any] ulid.ULID

func (d keyType[T]) Kind() string {
	var zero T
	return strings.TrimSuffix(reflect.TypeOf(zero).Name(), "Key")
}
func (d keyType[T]) String() string {
	return "ftl:" + d.Kind() + ":" + ulid.ULID(d).String()
}
func (d keyType[T]) ULID() ulid.ULID              { return ulid.ULID(d) }
func (d keyType[T]) MarshalText() ([]byte, error) { return []byte(d.String()), nil }
func (d *keyType[T]) UnmarshalText(bytes []byte) error {
	id, err := parseKey[keyType[T]](string(bytes))
	if err != nil {
		return err
	}
	*d = id
	return nil
}
