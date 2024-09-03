package ftl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/TBD54566975/ftl/go-runtime/encoding"
)

type Encrypted[T any] struct {
	isEncrypted bool
	data        []byte
}

var _ encoding.GenericMarshaler = (*Encrypted[int])(nil)
var _ encoding.GenericUnmarshaler = (*Encrypted[int])(nil)

func (e Encrypted[T]) Marshal(w *bytes.Buffer, encode func(v reflect.Value, w *bytes.Buffer) error) error {
	panic("unimplemented")
}

func (e Encrypted[T]) Unmarshal(d *json.Decoder, decode func(d *json.Decoder, v reflect.Value) error) error {
	var value T

	// Decode the JSON into the value
	if err := decode(d, reflect.ValueOf(&value).Elem()); err != nil {
		return fmt.Errorf("failed to decode: %w", err)
	}

	// Marshal the value into a []byte
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	e.data = data
	e.isEncrypted = false

	return nil
}
