// Package encoding defines the internal encoding that FTL uses to encode and
// decode messages. It is currently JSON.
package encoding

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"github.com/TBD54566975/ftl/backend/schema/strcase"
	"github.com/TBD54566975/ftl/go-runtime/ftl/typeregistry"
)

var (
	optionMarshaler   = reflect.TypeFor[OptionMarshaler]()
	optionUnmarshaler = reflect.TypeFor[OptionUnmarshaler]()
)

type OptionMarshaler interface {
	Marshal(ctx context.Context, w *bytes.Buffer, encode func(ctx context.Context, v reflect.Value, w *bytes.Buffer) error) error
}
type OptionUnmarshaler interface {
	Unmarshal(ctx context.Context, d *json.Decoder, isNull bool, decode func(ctx context.Context, d *json.Decoder, v reflect.Value) error) error
}

func Marshal(ctx context.Context, v any) ([]byte, error) {
	w := &bytes.Buffer{}
	err := encodeValue(ctx, reflect.ValueOf(v), w)
	return w.Bytes(), err
}

func encodeValue(ctx context.Context, v reflect.Value, w *bytes.Buffer) error {
	if !v.IsValid() {
		w.WriteString("null")
		return nil
	}

	if v.Kind() == reflect.Ptr {
		return fmt.Errorf("pointer types are not supported: %s", v.Type())
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

	case t.Implements(optionMarshaler):
		enc := v.Interface().(OptionMarshaler) //nolint:forcetypeassert
		return enc.Marshal(ctx, w, encodeValue)

	//TODO: Remove once we support `omitempty` tag
	case t == reflect.TypeFor[json.RawMessage]():
		data, err := json.Marshal(v.Interface())
		if err != nil {
			return err
		}
		w.Write(data)
		return nil
	}

	switch v.Kind() {
	case reflect.Struct:
		return encodeStruct(ctx, v, w)

	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return encodeBytes(v, w)
		}
		return encodeSlice(ctx, v, w)

	case reflect.Map:
		return encodeMap(ctx, v, w)

	case reflect.String:
		return encodeString(v, w)

	case reflect.Int:
		return encodeInt(v, w)

	case reflect.Float64:
		return encodeFloat(v, w)

	case reflect.Bool:
		return encodeBool(v, w)

	case reflect.Interface:
		if t == reflect.TypeFor[any]() {
			return encodeValue(ctx, v.Elem(), w)
		}

		if tr, ok := typeregistry.FromContext(ctx).Get(); ok {
			if vName, ok := tr.GetVariantByType(v.Type(), v.Elem().Type()).Get(); ok {
				sumType := struct {
					Name  string
					Value any
				}{Name: vName, Value: v.Elem().Interface()}
				return encodeValue(ctx, reflect.ValueOf(sumType), w)
			}
		}

		return fmt.Errorf("the only supported interface types are enums or any, not %s", t)

	default:
		panic(fmt.Sprintf("unsupported type: %s", v.Type()))
	}
}

func encodeStruct(ctx context.Context, v reflect.Value, w *bytes.Buffer) error {
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
		if err := encodeValue(ctx, fv, w); err != nil {
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

func encodeSlice(ctx context.Context, v reflect.Value, w *bytes.Buffer) error {
	w.WriteRune('[')
	for i := range v.Len() {
		if i > 0 {
			w.WriteRune(',')
		}
		if err := encodeValue(ctx, v.Index(i), w); err != nil {
			return err
		}
	}
	w.WriteRune(']')
	return nil
}

func encodeMap(ctx context.Context, v reflect.Value, w *bytes.Buffer) error {
	w.WriteRune('{')
	for i, key := range v.MapKeys() {
		if i > 0 {
			w.WriteRune(',')
		}
		w.WriteRune('"')
		w.WriteString(key.String())
		w.WriteString(`":`)
		if err := encodeValue(ctx, v.MapIndex(key), w); err != nil {
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

func Unmarshal(ctx context.Context, data []byte, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("unmarshal expects a non-nil pointer")
	}

	d := json.NewDecoder(bytes.NewReader(data))
	return decodeValue(ctx, d, rv.Elem())
}

func decodeValue(ctx context.Context, d *json.Decoder, v reflect.Value) error {
	if !v.CanSet() {
		return fmt.Errorf("cannot set value: %s", v.Type())
	}

	if v.Kind() == reflect.Ptr {
		return fmt.Errorf("pointer types are not supported: %s", v.Type())
	}

	t := v.Type()
	// Special-case types
	switch {
	case t == reflect.TypeFor[time.Time]():
		return d.Decode(v.Addr().Interface())

	case v.CanAddr() && v.Addr().Type().Implements(optionUnmarshaler):
		v = v.Addr()
		fallthrough

	case t.Implements(optionUnmarshaler):
		if v.IsNil() {
			v.Set(reflect.New(t.Elem()))
		}
		dec := v.Interface().(OptionUnmarshaler) //nolint:forcetypeassert
		return handleIfNextTokenIsNull(d, func(d *json.Decoder) error {
			return dec.Unmarshal(ctx, d, true, decodeValue)
		}, func(d *json.Decoder) error {
			return dec.Unmarshal(ctx, d, false, decodeValue)
		})
	}

	switch v.Kind() {
	case reflect.Struct:
		return decodeStruct(ctx, d, v)

	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return decodeBytes(d, v)
		}
		return decodeSlice(ctx, d, v)

	case reflect.Map:
		return decodeMap(ctx, d, v)

	case reflect.Interface:
		if tr, ok := typeregistry.FromContext(ctx).Get(); ok {
			if tr.IsSumTypeDiscriminator(v.Type()) {
				return decodeSumType(ctx, d, v)
			}
		}

		if v.Type().NumMethod() != 0 {
			return fmt.Errorf("the only supported interface types are enums or any, not %s", v.Type())
		}
		fallthrough

	default:
		return d.Decode(v.Addr().Interface())
	}
}

func decodeStruct(ctx context.Context, d *json.Decoder, v reflect.Value) error {
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
			if err := decodeValue(ctx, d, field); err != nil {
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

func decodeSlice(ctx context.Context, d *json.Decoder, v reflect.Value) error {
	if err := expectDelim(d, '['); err != nil {
		return err
	}

	for d.More() {
		newElem := reflect.New(v.Type().Elem()).Elem()
		if err := decodeValue(ctx, d, newElem); err != nil {
			return err
		}
		v.Set(reflect.Append(v, newElem))
	}
	// consume the closing delimiter of the slice
	_, err := d.Token()
	return err
}

func decodeMap(ctx context.Context, d *json.Decoder, v reflect.Value) error {
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
		if err := decodeValue(ctx, d, newElem); err != nil {
			return err
		}

		v.SetMapIndex(reflect.ValueOf(key), newElem)
	}
	// consume the closing delimiter of the map
	_, err := d.Token()
	return err
}

func decodeSumType(ctx context.Context, d *json.Decoder, v reflect.Value) error {
	tr, ok := typeregistry.FromContext(ctx).Get()
	if !ok {
		return fmt.Errorf("no type registry found in context")
	}

	var sumType struct {
		Name  string
		Value json.RawMessage
	}
	err := d.Decode(&sumType)
	if err != nil {
		return err
	}
	if sumType.Name == "" {
		return fmt.Errorf("no name found for type enum variant")
	}
	if sumType.Value == nil {
		return fmt.Errorf("no value found for type enum variant")
	}

	variantType, ok := tr.GetVariantByName(v.Type(), sumType.Name).Get()
	if !ok {
		return fmt.Errorf("no enum variant found by name %s", sumType.Name)
	}

	out := reflect.New(variantType)
	if err := decodeValue(ctx, json.NewDecoder(bytes.NewReader(sumType.Value)), out.Elem()); err != nil {
		return err
	}
	if !out.Type().AssignableTo(v.Type()) {
		return fmt.Errorf("cannot assign %s to %s", out.Type(), v.Type())
	}
	v.Set(out.Elem())

	return nil
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
//
// It reads the buffered data and checks if the next token is "null" without actually consuming the token.
func isNextTokenNull(d *json.Decoder) (bool, error) {
	s, err := io.ReadAll(d.Buffered())
	if err != nil {
		return false, err
	}

	remaining := s[bytes.IndexFunc(s, isDelim)+1:]
	secondDelim := bytes.IndexFunc(remaining, isDelim)
	if secondDelim == -1 {
		secondDelim = len(remaining) // No delimiters found, read until the end
	}

	return strings.TrimSpace(string(remaining[:secondDelim])) == "null", nil
}

func isDelim(r rune) bool {
	switch r {
	case ',', ':', '{', '}', '[', ']':
		return true
	}
	return false
}
