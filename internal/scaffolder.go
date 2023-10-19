package internal

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/alecthomas/errors"
	"github.com/iancoleman/strcase"
)

// Scaffold evaluates the scaffolding files at the given destination against
// ctx.
//
// Both paths and file contents are evaluated.
//
// The functions "snake", "camel", "lowerCamel", "kebab", "upper", and "lower"
// are available.
func Scaffold(destination string, ctx any) error {
	return errors.WithStack(walkDir(destination, func(path string, d fs.DirEntry) error {
		info, err := d.Info()
		if err != nil {
			return errors.WithStack(err)
		}

		// Evaluate path name templates.
		dir := filepath.Dir(path)
		base := filepath.Base(path)
		newName, err := evaluate(base, ctx)
		if err != nil {
			return errors.Wrapf(err, "%s", path)
		}
		// Rename if necessary.
		if newName != base {
			newName = filepath.Join(dir, newName)
			err = os.Rename(path, newName)
			if err != nil {
				return errors.Wrap(err, "failed to rename file")
			}
			path = newName
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		// Evaluate file content.
		template, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "%s", path)
		}
		content, err := evaluate(string(template), ctx)
		if err != nil {
			return errors.Wrapf(err, "%s", path)
		}
		err = os.WriteFile(path, []byte(content), info.Mode())
		if err != nil {
			return errors.Wrapf(err, "%s", path)
		}
		return nil
	}))
}

func walkDir(dir string, fn func(path string, d fs.DirEntry) error) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			err = walkDir(filepath.Join(dir, entry.Name()), fn)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		err = fn(filepath.Join(dir, entry.Name()), entry)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func evaluate(tmpl string, ctx any) (string, error) {
	t, err := template.New("scaffolding").Funcs(
		template.FuncMap{
			"snake":      strcase.ToSnake,
			"camel":      strcase.ToCamel,
			"lowerCamel": strcase.ToLowerCamel,
			"kebab":      strcase.ToKebab,
			"upper":      strings.ToUpper,
			"lower":      strings.ToLower,
			"typename": func(v any) string {
				return reflect.Indirect(reflect.ValueOf(v)).Type().Name()
			},
		},
	).Parse(tmpl)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse template")
	}
	newName := &strings.Builder{}
	err = t.Execute(newName, ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to execute template")
	}
	return newName.String(), nil
}
