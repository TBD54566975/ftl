package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/iancoleman/strcase"
	"golang.org/x/exp/slices"
)

func (s *Schema) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(marshalJSON(s))
	return data, errors.WithStack(err)
}

func (s *Schema) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, s)
	return errors.WithStack(err)
}

type jsonStruct struct {
	kind  string
	field []struct {
		name  string
		value any
	}
}

func (j *jsonStruct) MarshalJSON() ([]byte, error) {
	w := &bytes.Buffer{}
	w.WriteByte('{')
	fmt.Fprintf(w, `"kind":%q`, j.kind)
	if len(j.field) > 0 {
		for _, f := range j.field {
			w.WriteByte(',')
			fmt.Fprintf(w, `%q:`, f.name)
			if err := json.NewEncoder(w).Encode(f.value); err != nil {
				return nil, errors.WithStack(err)
			}
		}
	}
	w.WriteByte('}')
	return w.Bytes(), nil
}

func marshalJSON(v any) any {
	rv := reflect.Indirect(reflect.ValueOf(v))
	switch rv.Kind() {
	case reflect.Struct:
		out := &jsonStruct{
			kind: strcase.ToLowerCamel(rv.Type().Name()),
		}
		fields := reflect.VisibleFields(rv.Type())
		slices.SortFunc(fields, func(a, b reflect.StructField) bool { return a.Name < b.Name })
		for _, ft := range fields {
			jsonTag := strings.Split(ft.Tag.Get("json"), ",")
			if jsonTag[0] == "-" || !ft.IsExported() {
				continue
			}
			f := rv.FieldByIndex(ft.Index)
			if len(jsonTag) > 1 && jsonTag[1] == "omitempty" && f.IsZero() {
				continue
			}
			out.field = append(out.field, struct {
				name  string
				value any
			}{name: strcase.ToLowerCamel(ft.Name), value: marshalJSON(f.Interface())})
		}
		return out

	case reflect.Slice:
		out := []any{}
		for i := 0; i < rv.Len(); i++ {
			out = append(out, marshalJSON(rv.Index(i).Interface()))
		}
		return out

	case reflect.Map:
		out := map[string]any{}
		keys := rv.MapKeys()
		slices.SortFunc(keys, func(a, b reflect.Value) bool { return a.String() < b.String() })
		for _, k := range keys {
			out[k.String()] = marshalJSON(rv.MapIndex(k).Interface())
		}
		return out

	case reflect.Int:
		return rv.Int()

	case reflect.Float64:
		return rv.Float()

	case reflect.String:
		return rv.String()

	case reflect.Bool:
		return rv.Bool()

	default:
		panic("unhandled kind")
	}
}
