//nolint:wrapcheck
package ftl

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"

	ftlencoding "github.com/TBD54566975/ftl/go-runtime/encoding"
)

// Stdlib interfaces types implement.
type stdlib interface {
	fmt.Stringer
	fmt.GoStringer
}

// An Option type is a type that can contain a value or nothing.
type Option[T any] struct {
	value T
	ok    bool
}

var _ driver.Valuer = (*Option[int])(nil)
var _ sql.Scanner = (*Option[int])(nil)

func (o *Option[T]) Scan(src any) error {
	if src == nil {
		o.ok = false
		var zero T
		o.value = zero
		return nil
	}
	if value, ok := src.(T); ok {
		o.value = value
		o.ok = true
		return nil
	}
	var value T
	switch scan := any(&value).(type) {
	case sql.Scanner:
		if err := scan.Scan(src); err != nil {
			return fmt.Errorf("cannot scan %T into Option[%T]: %w", src, o.value, err)
		}
		o.value = value
		o.ok = true

	case encoding.TextUnmarshaler:
		switch src := src.(type) {
		case string:
			if err := scan.UnmarshalText([]byte(src)); err != nil {
				return fmt.Errorf("unmarshal from %T into Option[%T] failed: %w", src, o.value, err)
			}
			o.value = value
			o.ok = true

		case []byte:
			if err := scan.UnmarshalText(src); err != nil {
				return fmt.Errorf("cannot scan %T into Option[%T]: %w", src, o.value, err)
			}
			o.value = value
			o.ok = true

		default:
			return fmt.Errorf("cannot unmarshal %T into Option[%T]", src, o.value)
		}

	default:
		return fmt.Errorf("no decoding mechanism found for %T into Option[%T]", src, o.value)
	}
	return nil
}

func (o Option[T]) Value() (driver.Value, error) {
	if !o.ok {
		return nil, nil
	}
	switch value := any(o.value).(type) {
	case driver.Valuer:
		return value.Value()

	case encoding.TextMarshaler:
		return value.MarshalText()
	}
	return o.value, nil
}

var _ stdlib = (*Option[int])(nil)

// Some returns an Option that contains a value.
func Some[T any](value T) Option[T] { return Option[T]{value: value, ok: true} }

// None returns an Option that contains nothing.
func None[T any]() Option[T] { return Option[T]{} }

// Ptr returns an Option that is invalid if the pointer is nil, otherwise the dereferenced pointer.
func Ptr[T any](ptr *T) Option[T] {
	if ptr == nil {
		return None[T]()
	}
	return Some(*ptr)
}

// Nil returns an Option that is invalid if the value is nil, otherwise the value.
//
// If the type is not nillable (slice, map, chan, ptr, interface) this will panic.
func Nil[T any](ptr T) Option[T] {
	rv := reflect.ValueOf(ptr)
	if !rv.IsValid() {
		return None[T]()
	}
	switch rv.Type().Kind() {
	case reflect.Slice, reflect.Map, reflect.Chan, reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return None[T]()
		}
		return Some(ptr)

	default:
		panic("type is not nillable")
	}
}

// Zero returns an Option that is invalid if the value is the zero value, otherwise the value.
func Zero[T any](value T) Option[T] {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() || rv.IsZero() {
		return None[T]()
	}
	return Some(value)
}

// Ptr returns a pointer to the value if the Option contains a value, otherwise nil.
func (o Option[T]) Ptr() *T {
	if o.ok {
		return &o.value
	}
	return nil
}

// Ok returns true if the Option contains a value.
func (o Option[T]) Ok() bool { return o.ok }

// MustGet returns the value. It panics if the Option contains nothing.
func (o Option[T]) MustGet() T {
	if !o.ok {
		var t T
		panic(fmt.Sprintf("Option[%T] contains nothing", t))
	}
	return o.value
}

// Get returns the value and a boolean indicating if the Option contains a value.
func (o Option[T]) Get() (T, bool) { return o.value, o.ok }

// Default returns the Option value if it is present, otherwise it returns the
// value passed.
func (o Option[T]) Default(value T) T {
	if o.ok {
		return o.value
	}
	return value
}

func (o Option[T]) MarshalJSON() ([]byte, error) {
	if o.ok {
		return ftlencoding.Marshal(o.value)
	}
	return []byte("null"), nil
}

func (o *Option[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		o.ok = false
		return nil
	}
	if err := ftlencoding.Unmarshal(data, &o.value); err != nil {
		return err
	}
	o.ok = true
	return nil
}

func (o Option[T]) String() string {
	if o.ok {
		return fmt.Sprintf("%v", o.value)
	}
	return "None"
}

func (o Option[T]) GoString() string {
	if o.ok {
		return fmt.Sprintf("Some[%T](%#v)", o.value, o.value)
	}
	return fmt.Sprintf("None[%T]()", o.value)
}

func (o Option[T]) Marshal(w *bytes.Buffer, encode func(v reflect.Value, w *bytes.Buffer) error) error {
	if o.ok {
		return encode(reflect.ValueOf(&o.value).Elem(), w)
	}
	w.WriteString("null")
	return nil
}

func (o *Option[T]) Unmarshal(d *json.Decoder, isNull bool, decode func(d *json.Decoder, v reflect.Value) error) error {
	if isNull {
		o.ok = false
		return nil
	}
	if err := decode(d, reflect.ValueOf(&o.value).Elem()); err != nil {
		return err
	}
	o.ok = true
	return nil
}
