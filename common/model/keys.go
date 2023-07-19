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

func NewControllerKey() ControllerKey                      { return ControllerKey(ulid.Make()) }
func ParseControllerKey(key string) (ControllerKey, error) { return parseKey[ControllerKey](key) }

type controllerKey struct{}
type ControllerKey = keyType[controllerKey]

func NewIngressRequestKey() IngressRequestKey { return IngressRequestKey(ulid.Make()) }
func ParseIngressRequestKey(key string) (IngressRequestKey, error) {
	return parseKey[IngressRequestKey](key)
}

type ingressRequestKey struct{}
type IngressRequestKey = keyType[ingressRequestKey]

func parseKey[KT keyType[U], U any](key string) (KT, error) {
	var zero KT
	kind := kindFromType[U]()
	if !strings.HasPrefix(key, kind) {
		return zero, errors.Errorf("invalid %s key %q", kind, key)
	}
	ulid, err := ulid.Parse(key[len(kind):])
	if err != nil {
		return KT(ulid), err
	}
	return KT(ulid), nil
}

// Helper type to avoid having to write a bunch of boilerplate. It relies on T being a
// named struct in the form <name>Key, eg. "runnerKey"
type keyType[T any] ulid.ULID

func (d keyType[T]) Kind() string                 { return kindFromType[T]() }
func (d keyType[T]) String() string               { return d.Kind() + ulid.ULID(d).String() }
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

func kindFromType[T any]() string {
	var zero T
	return strings.ToUpper(strings.TrimSuffix(reflect.TypeOf(zero).Name(), "Key")[:1])
}
