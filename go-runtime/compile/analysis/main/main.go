package main

import (
	"golang.org/x/tools/go/analysis/unitchecker"

	"github.com/TBD54566975/ftl/go-runtime/compile/analysis/typealias"
)

func main() {
	unitchecker.Main(
		typealias.Extractor,
	)
}
