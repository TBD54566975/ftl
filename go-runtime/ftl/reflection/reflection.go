package reflection

import (
	"fmt"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
)

// Module returns the FTL module currently being executed.
func Module() string {
	// Look through the stack for the outermost FTL module.
	pcs := make([]uintptr, 1024)
	pcs = pcs[:runtime.Callers(1, pcs)]
	var module string
	for _, pc := range pcs {
		pkg := strings.Split(runtime.FuncForPC(pc).Name(), ".")[0]
		if strings.HasPrefix(pkg, "ftl/") {
			module = strings.Split(pkg, "/")[1]
		}
	}
	if module == "" {
		debug.PrintStack()
		panic("must be called from an FTL module")
	}
	return module
}

// TypeRef returns the Ref for a Go type.
//
// Panics if called with a type outside of FTL.
func TypeRef[T any]() Ref {
	var v T
	t := reflect.TypeOf(v)
	return goRefToFTLRef(t.PkgPath() + "." + t.Name())
}

// TypeRefFromValue returns the Ref for a Go value.
//
// The value must be a named type such as a struct, enum, or sum type.
func TypeRefFromValue(v any) Ref {
	t := reflect.TypeOf(v)
	return Ref{Module: moduleForType(t), Name: strcase.ToUpperCamel(t.Name())}
}

// FuncRef returns the Ref for a Go function.
//
// Panics if called with a function outside FTL.
func FuncRef(call any) Ref {
	ref := runtime.FuncForPC(reflect.ValueOf(call).Pointer()).Name()
	return goRefToFTLRef(ref)
}

// TypeFromValue reflects a schema.Type from a Go value.
//
// The passed value must be a pointer to a value of the desired type. This is to
// ensure that interface values aren't dereferenced automatically by the Go
// compiler.
func TypeFromValue[T any, TP interface{ *T }](v TP) schema.Type {
	return ReflectTypeToSchemaType(reflect.TypeOf(v).Elem())
}

var AllowAnyPackageForTesting = false

// goRefToFTLRef converts a Go reference path to an FTL reference.
//
// examples:
// ftl/modulename.Verb
// ftl/modulename/subpackage.Verb
func goRefToFTLRef(ref string) Ref {
	if !AllowAnyPackageForTesting && !strings.HasPrefix(ref, "ftl/") {
		panic(fmt.Sprintf("invalid reference %q, must start with ftl/ ", ref))
	}
	parts := strings.Split(ref, "/")

	moduleIdx := 1
	if AllowAnyPackageForTesting && parts[0] != "ftl" {
		moduleIdx = 0
	}

	// module is at idx 1 after "ftl", unless we're allowing any package
	module := strings.Split(parts[moduleIdx], ".")[0]

	// verb is always in last part, after the last "."
	// subpackage is not included in returned ref
	finalPackageComponents := strings.Split(parts[len(parts)-1], ".")
	verb := strcase.ToLowerCamel(finalPackageComponents[len(finalPackageComponents)-1])

	return Ref{Module: module, Name: verb}
}

// ReflectTypeToSchemaType returns the FTL schema for a Go reflect.Type.
func ReflectTypeToSchemaType(t reflect.Type) schema.Type {
	switch t.Kind() {
	case reflect.Struct:
		// Handle well-known types.
		if reflect.TypeFor[time.Time]() == t {
			return &schema.Time{}
		}
		// Check if it's a sum-type discriminator.
		if sumType, ok := GetDiscriminatorByVariant(t).Get(); ok {
			return refForType(sumType)
		}

		return refForType(t)

	case reflect.Slice:
		return &schema.Array{Element: ReflectTypeToSchemaType(t.Elem())}

	case reflect.Map:
		return &schema.Map{Key: ReflectTypeToSchemaType(t.Key()), Value: ReflectTypeToSchemaType(t.Elem())}

	case reflect.Bool:
		return &schema.Bool{}

	case reflect.String:
		if t.PkgPath() != "" { // Enum
			return &schema.Ref{
				Module: moduleForType(t),
				Name:   strcase.ToUpperCamel(t.Name()),
			}
		}
		return &schema.String{}

	case reflect.Int:
		if t.PkgPath() != "" { // Enum
			return &schema.Ref{
				Module: moduleForType(t),
				Name:   strcase.ToUpperCamel(t.Name()),
			}
		}
		return &schema.Int{}

	case reflect.Float64:
		return &schema.Float{}

	case reflect.Interface:
		if t.NumMethod() == 0 { // any
			return &schema.Any{}
		}
		// Check if it's a sum-type discriminator.
		if !IsSumTypeDiscriminator(t) {
			panic(fmt.Sprintf("unsupported interface type %s", t))
		}
		return refForType(t)

	default:
		panic(fmt.Sprintf("unsupported FTL type %s", t))
	}
}

// Return the FTL module for a type or panic if it's not an FTL type.
func moduleForType(t reflect.Type) string {
	module := t.PkgPath()
	if !AllowAnyPackageForTesting && !strings.HasPrefix(module, "ftl/") {
		panic(fmt.Sprintf("invalid reference %q, must start with ftl/ ", module))
	}
	parts := strings.Split(module, "/")
	return parts[len(parts)-1]
}

// refForType returns the schema.Ref for a Go type.
//
// This is not type checked in any way, so only valid types should be passed.
func refForType(t reflect.Type) *schema.Ref {
	module := moduleForType(t)
	return &schema.Ref{Module: module, Name: strcase.ToUpperCamel(t.Name())}
}
