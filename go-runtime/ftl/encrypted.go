package ftl

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/TBD54566975/ftl/go-runtime/encoding"
	"github.com/TBD54566975/ftl/internal/encryption"
	"github.com/alecthomas/types/optional"
)

type FLESubKey struct{}

func (FLESubKey) SubKey() string { return "fle" }

type Encrypted[T any] struct {
	isEncrypted bool
	plain       optional.Option[T]
	encrypted   optional.Option[encryption.EncryptedColumn[FLESubKey]]
}

var _ encoding.GenericMarshaler = (*Encrypted[int])(nil)
var _ encoding.GenericUnmarshaler = (*Encrypted[int])(nil)

func (e Encrypted[T]) Marshal(w *bytes.Buffer, encode func(v reflect.Value, w *bytes.Buffer) error) error {
	panic("unimplemented")
}

// Decode the value from the JSON decoder as is and don't encrypt it yet, because we don't have the key.
func (e Encrypted[T]) Unmarshal(d *json.Decoder, decode func(d *json.Decoder, v reflect.Value) error) error {
	var value T

	// // Decode the JSON into the value
	// if err := decode(d, reflect.ValueOf(&value).Elem()); err != nil {
	// 	return fmt.Errorf("failed to decode: %w", err)
	// }

	// // Marshal the value into a []byte
	// data, err := json.Marshal(value)
	// if err != nil {
	// 	return fmt.Errorf("failed to marshal: %w", err)
	// }
	//

	fmt.Printf("Unmarshal: %v\n", value)

	e.plain = optional.Some(value)
	e.isEncrypted = false

	return nil
}

var _ driver.Valuer = (*Encrypted[int])(nil)
var _ sql.Scanner = (*Encrypted[int])(nil)

// A proxy into the encrypted value.
// Will error if the value is not encrypted.
func (e *Encrypted[T]) Value() (driver.Value, error) {
	if !e.isEncrypted {
		return nil, fmt.Errorf("value is not encrypted")
	}

	value, err := e.encrypted.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted value: %w", err)
	}

	return value, nil
}

// A proxy into the encrypted value.
func (e *Encrypted[T]) Scan(src any) error {
	e.isEncrypted = true
	e.encrypted = optional.Some(encryption.EncryptedColumn[FLESubKey]{})
	err := e.encrypted.Scan(src)
	if err != nil {
		return fmt.Errorf("failed to scan encrypted value: %w", err)
	}

	return nil
}
