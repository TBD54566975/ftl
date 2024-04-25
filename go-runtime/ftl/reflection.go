package ftl

import (
	"runtime"
	"strings"
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
