package internal

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/alecthomas/errors"
	"github.com/iancoleman/strcase"
	"github.com/otiai10/copy"
)

// Scaffold copies the scaffolding files from the given source to the given
// destination, evaluating any templates against ctx in the process.
//
// Both paths and file contents are evaluated.
//
// The functions "snake", "camel", "lowerCamel", "kebab", "upper", and "lower"
// are available.
func Scaffold(source fs.FS, destination string, ctx any) error {
	err := copy.Copy(".", destination, copy.Options{FS: source, PermissionControl: copy.AddPermission(0600)})
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(filepath.WalkDir(destination, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return errors.WithStack(err)
		}

		// Evaluate path name templates.
		newName, err := evaluate(path, ctx)
		if err != nil {
			return errors.Wrapf(err, "%s", path)
		}
		// Rename if necessary.
		if newName != path {
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

func evaluate(tmpl string, ctx any) (string, error) {
	t, err := template.New("scaffolding").Funcs(
		template.FuncMap{
			"snake":      strcase.ToSnake,
			"camel":      strcase.ToCamel,
			"lowerCamel": strcase.ToLowerCamel,
			"kebab":      strcase.ToKebab,
			"upper":      strings.ToUpper,
			"lower":      strings.ToLower,
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
