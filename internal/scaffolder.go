package internal

import (
	"reflect"
	"strings"
	"text/template"

	"github.com/TBD54566975/scaffolder"
	"github.com/alecthomas/errors"
	"github.com/iancoleman/strcase"
)

// Scaffold evaluates the scaffolding files at the given destination against
// ctx.
func Scaffold(destination string, ctx any) error {
	return errors.WithStack(scaffolder.Scaffold(destination, ctx, scaffolder.Functions(template.FuncMap{
		"snake":          strcase.ToSnake,
		"screamingSnake": strcase.ToScreamingSnake,
		"camel":          strcase.ToCamel,
		"lowerCamel":     strcase.ToLowerCamel,
		"kebab":          strcase.ToKebab,
		"screamingKebab": strcase.ToScreamingKebab,
		"upper":          strings.ToUpper,
		"lower":          strings.ToLower,
		"title":          strings.Title,
		"typename": func(v any) string {
			return reflect.Indirect(reflect.ValueOf(v)).Type().Name()
		},
	})))
}
