package reflection

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/alecthomas/types/once"
)

type ReflectedDatabaseHandle struct {
	DBType string
	DB     *once.Handle[*sql.DB]

	// configs
	Name string
}

func Database[T any](dbname string, init func(ref Ref) *ReflectedDatabaseHandle) Registree {
	ref := Ref{
		Module: moduleForType(reflect.TypeFor[T]()),
		Name:   dbname,
	}
	return func(t *TypeRegistry) {
		t.databases[ref] = init(ref)
	}
}

func getDatabaseName[T any]() string {
	typ := reflect.TypeFor[T]()
	var config T
	if typ.Kind() == reflect.Ptr {
		config = reflect.New(typ.Elem()).Interface().(T) //nolint:forcetypeassert
	} else {
		config = reflect.New(typ).Elem().Interface().(T) //nolint:forcetypeassert
	}

	nameMethod := reflect.ValueOf(config).MethodByName("Name")
	if !nameMethod.IsValid() {
		panic(fmt.Sprintf("type %T must implement ftl.DatabaseConfig but does not have a Name() method", config))
	}
	nameResult := nameMethod.Call(nil)
	name, ok := nameResult[0].Interface().(string)
	if !ok {
		panic(fmt.Sprintf("Name() method of type %T must return a string, but returned %T", config,
			nameResult[0].Interface()))
	}
	return name
}
