package ftl

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

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

// FuncRef returns the Ref for a Go function.
//
// Panics if called with a function outside FTL.
func FuncRef(call any) Ref {
	ref := runtime.FuncForPC(reflect.ValueOf(call).Pointer()).Name()
	return goRefToFTLRef(ref)
}

func goRefToFTLRef(ref string) Ref {
	if !strings.HasPrefix(ref, "ftl/") {
		panic(fmt.Sprintf("invalid reference %q, must start with ftl/ ", ref))
	}
	parts := strings.Split(ref[strings.LastIndex(ref, "/")+1:], ".")
	return Ref{parts[len(parts)-2], strcase.ToLowerCamel(parts[len(parts)-1])}
}
