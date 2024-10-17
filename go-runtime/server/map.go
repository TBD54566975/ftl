package server

import (
	"fmt"
	"reflect"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
)

func MappedHandle[R ftl.ResourceMapper[From, To], From, To any](resource reflection.VerbResource) reflection.VerbResource {
	return func() reflect.Value {
		handle := resource()
		var instance R
		var wasSet bool
		val := reflect.ValueOf(&instance).Elem()
		for i := range val.NumField() {
			field := val.Field(i)
			if handle.Type().AssignableTo(field.Type()) {
				if field.CanSet() {
					handleValue := handle.Convert(field.Type())
					field.Set(handleValue)
					wasSet = true
					break
				}
				panic(fmt.Sprintf("mapper must contain a field of type %s that is settable", handle.Type().String()))
			}
		}
		if !wasSet {
			panic(fmt.Sprintf("mapper must contain a field of type %s", handle.Type().String()))
		}
		db := ftl.NewMappedHandle(instance)
		return reflect.ValueOf(db)
	}
}
