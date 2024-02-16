// Package automaxprocs sets GOMAXPROCS to match Linux container CPU quota.
package automaxprocs

import (
	"fmt"
	"os"

	"go.uber.org/automaxprocs/maxprocs"
)

func init() {
	_, err := maxprocs.Set()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ftl:warning: non-fatal error setting GOMAXPROCS: %v\n", err)
	}
}
