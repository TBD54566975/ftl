package internal

import (
	"archive/zip"
	"os"
	"strings"

	"github.com/TBD54566975/scaffolder"

	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/backend/schema/strcase"
)

// ScaffoldZip is a convenience function for scaffolding a zip archive with scaffolder.
func ScaffoldZip(source *zip.Reader, destination string, ctx any, options ...scaffolder.Option) error {
	tmpDir, err := os.MkdirTemp("", "scaffold-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)
	if err := UnzipDir(source, tmpDir); err != nil {
		return err
	}
	options = append(options, scaffolder.Functions(scaffoldFuncs))
	return scaffolder.Scaffold(tmpDir, destination, ctx, options...)
}

var scaffoldFuncs = scaffolder.FuncMap{
	"snake":          strcase.ToLowerSnake,
	"screamingSnake": strcase.ToUpperSnake,
	"camel":          strcase.ToUpperCamel,
	"lowerCamel":     strcase.ToLowerCamel,
	"strippedCamel":  strcase.ToUpperStrippedCamel,
	"kebab":          strcase.ToLowerKebab,
	"screamingKebab": strcase.ToUpperKebab,
	"upper":          strings.ToUpper,
	"lower":          strings.ToLower,
	"title":          strings.Title,
	"typename":       schema.TypeName,
}
