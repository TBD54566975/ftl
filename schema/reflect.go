package schema

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/alecthomas/errors"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// ReflectVerbIntoModule uses reflection to extract an FTL schema for a Verb
// and insert the schema into "module".
func ReflectVerbIntoModule[Req, Resp any](module *Module, fn func(context.Context, Req) (Resp, error)) error {
	name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	name = strings.TrimPrefix(name, module.Name+"/")
	name = strings.ReplaceAll(name, ".", "/")
	rt := reflect.TypeOf(fn)
	reqRef, req, err := reflectData(module.Name, rt.In(1))
	if err != nil {
		return errors.Wrap(err, name)
	}
	respRef, resp, err := reflectData(module.Name, rt.Out(0))
	if err != nil {
		return errors.Wrap(err, name)
	}
	verb := Verb{
		Name:     name,
		Request:  reqRef,
		Response: respRef,
	}
	module.Verbs = append(module.Verbs, verb)
	// Deduplicate types.
	data := make(map[string]Data, len(module.Data)+len(req)+len(resp))
	kf := func(d Data) string { return d.Name }
	copySliceIntoMap(data, module.Data, kf)
	copySliceIntoMap(data, req, kf)
	copySliceIntoMap(data, resp, kf)
	module.Data = maps.Values(data)
	slices.SortFunc(module.Data, func(l, r Data) bool { return l.Name < r.Name })
	return nil
}

func reflectData(module string, t reflect.Type) (DataRef, []Data, error) {
	if t.Kind() != reflect.Struct {
		return DataRef{}, nil, errors.Errorf("data structure must be a Go struct but is %s", t)
	}
	out := []Data{}
	data := Data{
		Name: strings.TrimPrefix(t.PkgPath()+"/"+t.Name(), module+"/"),
	}
	for i := 0; i < t.NumField(); i++ {
		tf := t.Field(i)
		field, children, err := reflectField(module, tf)
		if err != nil {
			return DataRef{}, nil, errors.Wrap(err, field.Name)
		}
		out = append(out, children...)
		data.Fields = append(data.Fields, field)
	}
	if strings.HasPrefix(t.PkgPath(), module) {
		out = append([]Data{data}, out...)
	}
	return DataRef{Name: data.Name}, out, nil
}

func reflectField(module string, ft reflect.StructField) (Field, []Data, error) {
	t, data, err := reflectType(module, ft.Type)
	if err != nil {
		return Field{}, nil, errors.Wrap(err, ft.Name)
	}
	return Field{Name: ft.Name, Type: t}, data, err
}

func reflectType(module string, t reflect.Type) (Type, []Data, error) {
	switch t.Kind() {
	case reflect.Int:
		return Int{Int: true}, nil, nil
	case reflect.String:
		return String{Str: true}, nil, nil
	case reflect.Bool:
		return Bool{Bool: true}, nil, nil
	case reflect.Slice:
		el, data, err := reflectType(module, t.Elem())
		if err != nil {
			return nil, nil, errors.Errorf("invalid slice element type %s", t.Elem())
		}
		return el, data, nil
	case reflect.Struct:
		ref, data, err := reflectData(module, t)
		if err != nil {
			return nil, nil, errors.WithStack(err)
		}
		return ref, data, nil
	default:
		return nil, nil, fmt.Errorf("type %s is not supported by FTL (yet)", t)
	}
}

func copySliceIntoMap[K comparable, V any](out map[K]V, slice []V, kf func(V) K) {
	for _, el := range slice {
		out[kf(el)] = el
	}
}
