//nolint:revive
package model

import (
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"fmt"
	"reflect"
	"strings"

	base36 "github.com/multiformats/go-base36"
)

// Overridable random source for testing
var randRead = rand.Read

// A constraint that requires itself be a pointer to a T that implements KeyPayload.
//
// This is necessary so that keyType.Payload can be a value rather than a pointer.
type keyPayloadConstraint[T any] interface {
	*T
	KeyPayload
}

// KeyPayload is an interface that all key payloads must implement.
type KeyPayload interface {
	Kind() string
	String() string
	// Parse the hyphen-separated parts of the payload
	Parse(parts []string) error
	// RandomBytes determines the number of random bytes the key should include.
	RandomBytes() int
}

// KeyType is a helper type to avoid having to write a bunch of boilerplate.
type KeyType[T any, TP keyPayloadConstraint[T]] struct {
	Payload T
	Suffix  []byte
}

var _ interface {
	sql.Scanner
	driver.Valuer
	encoding.TextUnmarshaler
	encoding.TextMarshaler
} = (*KeyType[ControllerPayload, *ControllerPayload])(nil)

func (d KeyType[T, TP]) IsZero() bool {
	return d.Equal(KeyType[T, TP]{})
}

func (d KeyType[T, TP]) Equal(other KeyType[T, TP]) bool {
	return reflect.DeepEqual(d, other)
}

func (d KeyType[T, TP]) Value() (driver.Value, error) {
	return d.String(), nil
}

// Scan from DB representation.
func (d *KeyType[T, TP]) Scan(src any) error {
	input, ok := src.(string)
	if !ok {
		return fmt.Errorf("expected key to be a string but it's a %T", src)
	}
	key, err := parseKey[T, TP](input)
	if err != nil {
		return err
	}
	*d = key
	return nil
}

func (d KeyType[T, TP]) Kind() string {
	var payload TP = &d.Payload
	return payload.Kind()
}

func (d KeyType[T, TP]) String() string {
	parts := []string{d.Kind()}
	var payload TP = &d.Payload
	if payload := payload.String(); payload != "" {
		parts = append(parts, payload)
	}
	parts = append(parts, base36.EncodeToStringLc(d.Suffix))
	return strings.Join(parts, "-")
}

func (d KeyType[T, TP]) MarshalText() ([]byte, error) { return []byte(d.String()), nil }
func (d *KeyType[T, TP]) UnmarshalText(bytes []byte) error {
	id, err := parseKey[T, TP](string(bytes))
	if err != nil {
		return err
	}
	*d = id
	return nil
}

// Generate a new key.
//
// If the payload specifies a randomness greater than 0, a random suffix will be generated.
// The payload will be parsed from payloadComponents, which must be a hyphen-separated string.
func newKey[T any, TP keyPayloadConstraint[T]](components ...string) (kt KeyType[T, TP]) {
	var payload TP = &kt.Payload
	if err := payload.Parse(components); err != nil {
		panic(fmt.Errorf("failed to parse payload %q: %w", strings.Join(components, "-"), err))
	}
	if randomness := payload.RandomBytes(); randomness > 0 {
		kt.Suffix = make([]byte, randomness)
		if _, err := randRead(kt.Suffix); err != nil {
			panic(fmt.Errorf("failed to generate random suffix: %w", err))
		}
	}
	return kt
}

// Parse a key in the form <kind>[-<payload>][-<suffix>]
//
// Suffix will be parsed if the payload specifies a randomness greater than 0.
func parseKey[T any, TP keyPayloadConstraint[T]](key string) (kt KeyType[T, TP], err error) {
	components := strings.Split(key, "-")
	if len(components) == 0 {
		return kt, fmt.Errorf("expected a prefix for key %q", key)
	}

	// Validate and strip kind.
	var payload TP = &kt.Payload
	if components[0] != payload.Kind() {
		return kt, fmt.Errorf("expected prefix %q for key %q", payload.Kind(), key)
	}
	components = components[1:]

	// Optionally parse and strip random suffix.
	randomness := payload.RandomBytes()
	if randomness > 0 {
		if len(components) == 0 {
			return kt, fmt.Errorf("expected a suffix for key %q", key)
		}
		var err error
		kt.Suffix, err = base36.DecodeString(components[len(components)-1])
		if err != nil {
			return kt, fmt.Errorf("expected a base36 suffix for key %q: %w", key, err)
		}
		if len(kt.Suffix) != randomness {
			return kt, fmt.Errorf("expected a suffix of %d bytes for key %q, not %d", randomness, key, len(kt.Suffix))
		}
		components = components[:len(components)-1]
	}

	if err := payload.Parse(components); err != nil {
		return kt, fmt.Errorf("failed to parse payload for key %q: %w", key, err)
	}

	return kt, nil
}
