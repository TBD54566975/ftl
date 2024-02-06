// Package encoding defines the internal encoding that FTL uses to encode and
// decode messages. It is currently JSON.
package encoding

import (
	"bytes"
	"encoding"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/iancoleman/strcase"
)

var (
	textUnarmshaler = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	jsonUnmarshaler = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
)

func Marshal(v any) ([]byte, error) {
	w := &bytes.Buffer{}
	err := encodeValue(reflect.ValueOf(v), w)
	return w.Bytes(), err
}

func encodeValue(v reflect.Value, w *bytes.Buffer) error {
	t := v.Type()
	switch {
	case t.Kind() == reflect.Ptr && t.Elem().Implements(jsonUnmarshaler):
		v = v.Elem()
		fallthrough
	case t.Implements(jsonUnmarshaler):
		enc := v.Interface().(json.Marshaler) //nolint:forcetypeassert
		data, err := enc.MarshalJSON()
		if err != nil {
			return err
		}
		w.Write(data)
		return nil

	case t.Kind() == reflect.Ptr && t.Elem().Implements(textUnarmshaler):
		v = v.Elem()
		fallthrough
	case t.Implements(textUnarmshaler):
		enc := v.Interface().(encoding.TextMarshaler) //nolint:forcetypeassert
		data, err := enc.MarshalText()
		if err != nil {
			return err
		}
		data, err = json.Marshal(string(data))
		if err != nil {
			return err
		}
		w.Write(data)
		return nil
	}

	switch v.Kind() {
	case reflect.Struct:
		return encodeStruct(v, w)

	case reflect.Ptr:
		if v.IsNil() {
			w.WriteString("null")
			return nil
		}
		return encodeValue(v.Elem(), w)

	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return encodeBytes(v, w)
		}
		return encodeSlice(v, w)

	case reflect.Map:
		return encodeMap(v, w)

	case reflect.String:
		return encodeString(v, w)

	case reflect.Int:
		return encodeInt(v, w)

	case reflect.Float64:
		return encodeFloat(v, w)

	case reflect.Bool:
		return encodeBool(v, w)

	default:
		panic(fmt.Sprintf("unsupported typefoo: %s", v.Type()))
	}
}

func encodeStruct(v reflect.Value, w *bytes.Buffer) error {
	w.WriteRune('{')
	for i := 0; i < v.NumField(); i++ {
		if i > 0 {
			w.WriteRune(',')
		}
		field := v.Type().Field(i)
		w.WriteString(`"` + strcase.ToLowerCamel(field.Name) + `":`)
		if err := encodeValue(v.Field(i), w); err != nil {
			return err
		}
	}
	w.WriteRune('}')
	return nil
}

func encodeBytes(v reflect.Value, w *bytes.Buffer) error {
	w.WriteRune('"')
	data := base64.StdEncoding.EncodeToString(v.Bytes())
	w.WriteString(data)
	w.WriteRune('"')
	return nil
}

func encodeSlice(v reflect.Value, w *bytes.Buffer) error {
	w.WriteRune('[')
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			w.WriteRune(',')
		}
		if err := encodeValue(v.Index(i), w); err != nil {
			return err
		}
	}
	w.WriteRune(']')
	return nil
}

func encodeMap(v reflect.Value, w *bytes.Buffer) error {
	w.WriteRune('{')
	for i, key := range v.MapKeys() {
		if i > 0 {
			w.WriteRune(',')
		}
		w.WriteRune('"')
		w.WriteString(key.String())
		w.WriteString(`":`)
		if err := encodeValue(v.MapIndex(key), w); err != nil {
			return err
		}
	}
	w.WriteRune('}')
	return nil
}

func encodeBool(v reflect.Value, w *bytes.Buffer) error {
	if v.Bool() {
		w.WriteString("true")
	} else {
		w.WriteString("false")
	}
	return nil
}

func encodeInt(v reflect.Value, w *bytes.Buffer) error {
	fmt.Fprintf(w, "%d", v.Int())
	return nil
}

func encodeFloat(v reflect.Value, w *bytes.Buffer) error {
	fmt.Fprintf(w, "%g", v.Float())
	return nil
}

func encodeString(v reflect.Value, w *bytes.Buffer) error {
	w.WriteRune('"')
	fmt.Fprintf(w, "%s", v.String())
	w.WriteRune('"')
	return nil
}
