package schema

import (
	"strings"

	"github.com/alecthomas/errors"
)

// Validate performs semantic validation of a schema.
func Validate(schema Schema) error {
	modules := map[string]bool{}
	verbs := map[string]bool{}
	data := map[string]bool{}
	verbRefs := []VerbRef{}
	dataRefs := []DataRef{}
	merr := []error{}
	for _, module := range schema.Modules {
		_, seen := modules[module.Name]
		if seen {
			merr = append(merr, errors.Errorf("%s: duplicate module %q", module.Pos, module.Name))
		}
		modules[module.Name] = true
		if err := ValidateModule(module); err != nil {
			merr = append(merr, err)
		}
		err := Visit(module, func(n Node, next func() error) error {
			switch n := n.(type) {
			case VerbRef:
				verbRefs = append(verbRefs, n)
			case DataRef:
				dataRefs = append(dataRefs, n)
			case Verb:
				if n.Name == "" {
					merr = append(merr, errors.Errorf("%s: verb name is required", n.Pos))
					break
				}
				verbs[makeRef(module.Name, n.Name)] = true
				verbs[n.Name] = true
			case Data:
				if n.Name == "" {
					merr = append(merr, errors.Errorf("%s: data structure name is required", n.Pos))
					break
				}
				data[makeRef(module.Name, n.Name)] = true
				data[n.Name] = true
			default:
			}
			return next()
		})
		if err != nil {
			merr = append(merr, err)
		}
	}
	for _, ref := range verbRefs {
		if !verbs[ref.String()] {
			merr = append(merr, errors.Errorf("%s: reference to unknown Verb %q", ref.Pos, ref))
		}
	}
	for _, ref := range dataRefs {
		if !data[ref.String()] {
			merr = append(merr, errors.Errorf("%s: reference to unknown Data structure %q", ref.Pos, ref))
		}
	}
	return errors.Join(merr...)
}

// ValidateModule performs the subset of semantic validation possible on a single module.
func ValidateModule(module Module) error {
	merr := []error{}
	if module.Name == "" {
		merr = append(merr, errors.Errorf("%s: module name is required", module.Pos))
	}
	_ = Visit(module, func(n Node, next func() error) error {
		if n, ok := n.(Data); ok {
			for _, md := range n.Metadata {
				if md, ok := md.(MetadataCalls); ok {
					merr = append(merr, errors.Errorf("%s: metadata %q is not valid on data structures", md.Pos, strings.TrimSpace(md.String())))
				}
			}
		}
		return next()
	})
	return errors.Join(merr...)
}

func makeRef(module, name string) string {
	if module == "" {
		return name
	}
	return module + "." + name
}
