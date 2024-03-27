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
	"strings"

	"github.com/TBD54566975/ftl/backend/schema/strcase"
)

var (
	textMarshaler   = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
	textUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	jsonMarshaler   = reflect.TypeOf((*json.Marshaler)(nil)).Elem()
	jsonUnmarshaler = reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()
)

func Marshal(v any) ([]byte, error) {
	w := &bytes.Buffer{}
	err := encodeValue(reflect.ValueOf(v), w)
	return w.Bytes(), err
}

func encodeValue(v reflect.Value, w *bytes.Buffer) error {
	if !v.IsValid() {
		w.WriteString("null")
		return nil
	}
	t := v.Type()
	switch {
	case t.Kind() == reflect.Ptr && t.Elem().Implements(jsonMarshaler):
		v = v.Elem()
		fallthrough

	case t.Implements(jsonMarshaler):
		enc := v.Interface().(json.Marshaler) //nolint:forcetypeassert
		data, err := enc.MarshalJSON()
		if err != nil {
			return err
		}
		w.Write(data)
		return nil

	case t.Kind() == reflect.Ptr && t.Elem().Implements(textMarshaler):
		v = v.Elem()
		fallthrough

	case t.Implements(textMarshaler):
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

	case reflect.Interface: // any
		if t != reflect.TypeOf((*any)(nil)).Elem() {
			return fmt.Errorf("the only interface type supported is any, not %s", t)
		}
		return encodeValue(v.Elem(), w)

	default:
		panic(fmt.Sprintf("unsupported type: %s", v.Type()))
	}
}

func encodeStruct(v reflect.Value, w *bytes.Buffer) error {
	w.WriteRune('{')
	afterFirst := false
	for i := range v.NumField() {
		ft := v.Type().Field(i)
		t := ft.Type
		fv := v.Field(i)
		// Types that can be skipped if they're zero.
		if (t.Kind() == reflect.Slice && fv.Len() == 0) ||
			(t.Kind() == reflect.Map && fv.Len() == 0) ||
			(t.String() == "ftl.Unit" && fv.IsZero()) ||
			(strings.HasPrefix(t.String(), "ftl.Option[") && fv.IsZero()) ||
			(t == reflect.TypeOf((*any)(nil)).Elem() && fv.IsZero()) {
			continue
		}
		if afterFirst {
			w.WriteRune(',')
		}
		afterFirst = true
		w.WriteString(`"` + strcase.ToLowerCamel(ft.Name) + `":`)
		if err := encodeValue(fv, w); err != nil {
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
	for i := range v.Len() {
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

func Unmarshal(data []byte, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("unmarshal expects a non-nil pointer")
	}

	d := json.NewDecoder(bytes.NewReader(data))
	return decodeValue(d, rv.Elem())
}

func decodeValue(d *json.Decoder, v reflect.Value) error {
	if !v.CanSet() {
		return fmt.Errorf("cannot set value")
	}

	t := v.Type()
	switch {
	case v.Kind() != reflect.Ptr && v.CanAddr() && v.Addr().Type().Implements(jsonUnmarshaler):
		v = v.Addr()
		fallthrough

	case t.Implements(jsonUnmarshaler):
		if v.IsNil() {
			v.Set(reflect.New(t.Elem()))
		}
		o := v.Interface()
		return d.Decode(&o)

	case v.Kind() != reflect.Ptr && v.CanAddr() && v.Addr().Type().Implements(textUnmarshaler):
		v = v.Addr()
		fallthrough

	case t.Implements(textUnmarshaler):
		if v.IsNil() {
			v.Set(reflect.New(t.Elem()))
		}
		dec := v.Interface().(encoding.TextUnmarshaler) //nolint:forcetypeassert
		var s string
		if err := d.Decode(&s); err != nil {
			return err
		}
		return dec.UnmarshalText([]byte(s))
	}

	switch v.Kind() {
	case reflect.Struct:
		return decodeStruct(d, v)

	case reflect.Ptr:
		if token, err := d.Token(); err != nil {
			return err
		} else if token == nil {
			v.Set(reflect.Zero(v.Type()))
			return nil
		}
		return decodeValue(d, v.Elem())

	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return decodeBytes(d, v)
		}
		return decodeSlice(d, v)

	case reflect.Map:
		return decodeMap(d, v)

	case reflect.Interface:
		if v.Type().NumMethod() != 0 {
			return fmt.Errorf("the only interface type supported is any, not %s", v.Type())
		}
		fallthrough

	default:
		return d.Decode(v.Addr().Interface())
	}
}

func decodeStruct(d *json.Decoder, v reflect.Value) error {
	if err := expectDelim(d, '{'); err != nil {
		return err
	}

	for d.More() {
		token, err := d.Token()
		if err != nil {
			return err
		}
		key, ok := token.(string)
		if !ok {
			return fmt.Errorf("expected string key, got %T", token)
		}

		field := v.FieldByNameFunc(func(s string) bool {
			return strcase.ToLowerCamel(s) == key
		})
		if !field.IsValid() {
			return fmt.Errorf("no field corresponding to key %s", key)
		}
		fieldTypeStr := field.Type().String()
		switch {
		case fieldTypeStr == "*Unit" || fieldTypeStr == "Unit":
			if fieldTypeStr == "*Unit" && field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
		default:
			if err := decodeValue(d, field); err != nil {
				return err
			}
		}
	}

	// consume the closing delimiter of the object
	_, err := d.Token()
	return err
}

func decodeBytes(d *json.Decoder, v reflect.Value) error {
	var b []byte
	if err := d.Decode(&b); err != nil {
		return err
	}
	v.SetBytes(b)
	return nil
}

func decodeSlice(d *json.Decoder, v reflect.Value) error {
	if err := expectDelim(d, '['); err != nil {
		return err
	}

	for d.More() {
		newElem := reflect.New(v.Type().Elem()).Elem()
		if err := decodeValue(d, newElem); err != nil {
			return err
		}
		v.Set(reflect.Append(v, newElem))
	}
	// consume the closing delimiter of the slice
	_, err := d.Token()
	return err
}

func decodeMap(d *json.Decoder, v reflect.Value) error {
	if err := expectDelim(d, '{'); err != nil {
		return err
	}

	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	valType := v.Type().Elem()
	for d.More() {
		key, err := d.Token()
		if err != nil {
			return err
		}

		newElem := reflect.New(valType).Elem()
		if err := decodeValue(d, newElem); err != nil {
			return err
		}

		v.SetMapIndex(reflect.ValueOf(key), newElem)
	}
	// consume the closing delimiter of the map
	_, err := d.Token()
	return err
}

func expectDelim(d *json.Decoder, expected json.Delim) error {
	token, err := d.Token()
	if err != nil {
		return err
	}
	delim, ok := token.(json.Delim)
	if !ok || delim != expected {
		return fmt.Errorf("expected delimiter %q, got %q", expected, token)
	}
	return nil
}
