// Package encoding defines the internal encoding that FTL uses to encode and
// decode messages. It is currently JSON.
package encoding

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"
	"unicode"

	"github.com/TBD54566975/ftl/backend/schema/strcase"
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

	// Special-cased types
	switch {
	case t == reflect.TypeFor[time.Time]():
		data, err := json.Marshal(v.Interface())
		if err != nil {
			return err
		}
		w.Write(data)
		return nil

	case t == reflect.TypeFor[json.RawMessage]():
		data, err := json.Marshal(v.Interface())
		if err != nil {
			return err
		}
		w.Write(data)
		return nil

	case isOption(v.Type()):
		return encodeOption(v, w)
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

var ftlOptionTypePath = "github.com/TBD54566975/ftl/go-runtime/ftl.Option"

func isOption(t reflect.Type) bool {
	return strings.HasPrefix(t.PkgPath()+"."+t.Name(), ftlOptionTypePath)
}

func encodeOption(v reflect.Value, w *bytes.Buffer) error {
	if v.NumField() != 2 {
		return fmt.Errorf("value cannot have type ftl.Option since it has %d fields rather than 2: %v", v.NumField(), v)
	}
	optionOk := v.Field(1).Bool()
	if !optionOk {
		w.WriteString("null")
		return nil
	}
	return encodeValue(v.Field(0), w)
}

func encodeStruct(v reflect.Value, w *bytes.Buffer) error {
	w.WriteRune('{')
	afterFirst := false
	for i := range v.NumField() {
		ft := v.Type().Field(i)
		fv := v.Field(i)
		// TODO: If these fields are skipped, the ingress encoder will not include
		// them in the output. There should ideally be no relationship between
		// the ingress encoder and the encoding package, but for now this is the
		// simplest solution.

		// t := ft.Type
		// // Types that can be skipped if they're zero.
		// if (t.Kind() == reflect.Slice && fv.Len() == 0) ||
		// 	(t.Kind() == reflect.Map && fv.Len() == 0) ||
		// 	(t.String() == "ftl.Unit" && fv.IsZero()) ||
		// 	(strings.HasPrefix(t.String(), "ftl.Option[") && fv.IsZero()) ||
		// 	(t == reflect.TypeOf((*any)(nil)).Elem() && fv.IsZero()) {
		// 	continue
		// }
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
	data := base64.StdEncoding.EncodeToString(v.Bytes())
	fmt.Fprintf(w, "%q", data)
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
	fmt.Fprintf(w, "%q", v.String())
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
		allBytes, _ := io.ReadAll(d.Buffered())
		return fmt.Errorf("cannot set value: %v", string(allBytes))
	}

	t := v.Type()

	// Special-case types
	switch {
	case t == reflect.TypeFor[time.Time]():
		return d.Decode(v.Addr().Interface())
	case isOption(v.Type()):
		return decodeOption(d, v)
	}

	switch v.Kind() {
	case reflect.Struct:
		return decodeStruct(d, v)

	case reflect.Ptr:
		return handleIfNextTokenIsNull(d, func(d *json.Decoder) error {
			v.Set(reflect.Zero(v.Type()))
			return nil
		}, func(d *json.Decoder) error {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			return decodeValue(d, v.Elem())
		})

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

func handleIfNextTokenIsNull(d *json.Decoder, ifNullFn func(*json.Decoder) error, elseFn func(*json.Decoder) error) error {
	isNull, err := isNextTokenNull(d)
	if err != nil {
		return err
	}
	if isNull {
		err = ifNullFn(d)
		if err != nil {
			return err
		}
		// Consume the null token
		_, err := d.Token()
		if err != nil {
			return err
		}
		return nil
	}
	return elseFn(d)
}

// isNextTokenNull implements a cheap/dirty version of `Peek()`, which json.Decoder does
// not support.
func isNextTokenNull(d *json.Decoder) (bool, error) {
	s, err := io.ReadAll(d.Buffered())
	if err != nil {
		return false, err
	}
	if len(s) == 0 {
		return false, fmt.Errorf("cannot check emptystring for token \"null\"")
	}
	if s[0] != ':' {
		return false, fmt.Errorf("cannot check emptystring for token \"null\"")
	}
	i := 1
	for len(s) > i && unicode.IsSpace(rune(s[i])) {
		i++
	}
	if len(s) < i+4 {
		return false, nil
	}
	return string(s[i:i+4]) == "null", nil
}

func decodeOption(d *json.Decoder, v reflect.Value) error {
	return handleIfNextTokenIsNull(d, func(d *json.Decoder) error {
		v.FieldByName("Okay").SetBool(false)
		return nil
	}, func(d *json.Decoder) error {
		err := decodeValue(d, v.FieldByName("Val"))
		if err != nil {
			return err
		}
		v.FieldByName("Okay").SetBool(true)
		return nil
	})
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
