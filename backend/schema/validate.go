package schema

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/alecthomas/participle/v2"
	"golang.design/x/reflect"
	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/internal/errors"
)

var (
	// Primitive types can't be used as identifiers.
	//
	// We don't need Array/Map/VerbRef/DataRef here because there are no
	// keywords associated with these types.
	primitivesScope = Scope{
		"Int":    ModuleDecl{Decl: &Int{}},
		"Float":  ModuleDecl{Decl: &Float{}},
		"String": ModuleDecl{Decl: &String{}},
		"Bytes":  ModuleDecl{Decl: &Bytes{}},
		"Bool":   ModuleDecl{Decl: &Bool{}},
		"Time":   ModuleDecl{Decl: &Time{}},
		"Unit":   ModuleDecl{Decl: &Unit{}},
		"Any":    ModuleDecl{Decl: &Any{}},
	}
)

// MustValidate panics if a schema is invalid.
//
// This is useful for testing.
func MustValidate(schema *Schema) *Schema {
	clone, err := Validate(schema)
	if err != nil {
		panic(err)
	}
	return clone
}

// Validate clones, normalises and semantically valies a schema.
func Validate(schema *Schema) (*Schema, error) {
	schema = reflect.DeepCopy(schema)
	modules := map[string]bool{}
	merr := []error{}
	ingress := map[string]*Verb{}

	// Inject builtins.
	builtins := Builtins()
	// Move builtins to the front of the list.
	schema.Modules = slices.DeleteFunc(schema.Modules, func(m *Module) bool { return m.Name == builtins.Name })
	schema.Modules = append([]*Module{builtins}, schema.Modules...)

	scopes := NewScopes()

	// First pass, add all the modules.
	for _, module := range schema.Modules {
		if module == builtins {
			continue
		}
		if err := scopes.Add(nil, module.Name, module); err != nil {
			merr = append(merr, err)
		}
	}

	// Validate modules.
	for _, module := range schema.Modules {
		// Skip builtin module, it's already been validated.
		if module.Name == "builtin" {
			continue
		}

		if _, seen := modules[module.Name]; seen {
			merr = append(merr, fmt.Errorf("%s: duplicate module %q", module.Pos, module.Name))
		}
		modules[module.Name] = true
		if err := ValidateModule(module); err != nil {
			merr = append(merr, err)
		}

		indent := 0
		err := Visit(module, func(n Node, next func() error) error {
			save := scopes
			if scoped, ok := n.(Scoped); ok {
				scopes = scopes.PushScope(scoped.Scope())
				defer func() { scopes = save }()
			}
			indent++
			defer func() { indent-- }()
			switch n := n.(type) {
			case *VerbRef:
				if mdecl := scopes.Resolve(n.Untyped()); mdecl != nil {
					if _, ok := mdecl.Decl.(*Verb); !ok {
						merr = append(merr, fmt.Errorf("%s: reference to invalid verb %q at %q", n.Pos, n, mdecl.Decl.Position()))
					} else if mdecl.Module != nil {
						n.Module = mdecl.Module.Name
					}
				} else {
					merr = append(merr, fmt.Errorf("%s: reference to unknown verb %q", n.Pos, n))
				}

			case *DataRef:
				if mdecl := scopes.Resolve(n.Untyped()); mdecl != nil {
					switch decl := mdecl.Decl.(type) {
					case *Data:
						if mdecl.Module != nil {
							n.Module = mdecl.Module.Name
						}
						if len(n.TypeParameters) != len(decl.TypeParameters) {
							merr = append(merr, fmt.Errorf("%s: reference to data structure %s has %d type parameters, but %d were expected",
								n.Pos, n.Name, len(n.TypeParameters), len(decl.TypeParameters)))
						}

					case *TypeParameter:

					default:
						merr = append(merr, fmt.Errorf("%s: reference to invalid data structure %q at %s", n.Pos, n, mdecl.Decl.Position()))
					}
				} else {
					merr = append(merr, fmt.Errorf("%s: reference to unknown data structure %q", n.Pos, n))
				}

			case *Verb:
				for _, md := range n.Metadata {
					if md, ok := md.(*MetadataIngress); ok {
						if existing, ok := ingress[md.String()]; ok {
							merr = append(merr, fmt.Errorf("%s: duplicate %q for %s:%q and %s:%q", md.Pos, md.String(), existing.Pos, existing.Name, n.Pos, n.Name))
						}

						if md.Type == "http" && (!strings.HasPrefix(n.Request.String(), "builtin.HttpRequest") || !strings.HasPrefix(n.Response.String(), "builtin.HttpResponse")) {
							merr = append(merr, fmt.Errorf("%s: HTTP ingress verb %s(%s) %s must have the signature %s(builtin.HttpRequest) builtin.HttpResponse",
								md.Pos, n.Name, n.Request, n.Response, n.Name))
						}
						ingress[md.String()] = n
					}
				}

			case *Array, *Bool, *Bytes, *Data, *Database, Decl, *Field, *Float,
				IngressPathComponent, *IngressPathLiteral, *IngressPathParameter,
				*Int, *Map, Metadata, *MetadataCalls, *MetadataDatabases,
				*MetadataIngress, *Module, *Optional, *Schema, *String, *Time, Type,
				*Unit, *Any, *TypeParameter:
			}
			return next()
		})
		if err != nil {
			merr = append(merr, err)
		}
	}

	merr = cleanErrors(merr)
	return schema, errors.Join(merr...)
}

var validNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ValidateName validates an FTL name.
func ValidateName(name string) bool {
	return validNameRe.MatchString(name)
}

// ValidateModule performs the subset of semantic validation possible on a single module.
//
// It ignores references to other modules.
func ValidateModule(module *Module) error {
	merr := []error{}

	scopes := NewScopes()

	if !ValidateName(module.Name) {
		merr = append(merr, fmt.Errorf("%s: module name %q is invalid", module.Pos, module.Name))
	}
	if module.Builtin && module.Name != "builtin" {
		merr = append(merr, fmt.Errorf("%s: only the \"ftl\" module can be marked as builtin", module.Pos))
	}
	if err := scopes.Add(nil, module.Name, module); err != nil {
		merr = append(merr, err)
	}
	scopes = scopes.Push()
	_ = Visit(module, func(n Node, next func() error) error {
		if scoped, ok := n.(Scoped); ok {
			pop := scopes
			scopes = scopes.PushScope(scoped.Scope())
			defer func() { scopes = pop }()
		}
		switch n := n.(type) {
		case *VerbRef:
			if mdecl := scopes.Resolve(n.Untyped()); mdecl != nil {
				if _, ok := mdecl.Decl.(*Verb); !ok && n.Module == "" {
					merr = append(merr, fmt.Errorf("%s: unqualified reference to invalid verb %q", n.Pos, n))
				} else {
					n.Module = mdecl.Module.Name
				}
			} else if n.Module == "" || n.Module == module.Name { // Don't report errors for external modules.
				merr = append(merr, fmt.Errorf("%s: reference to unknown verb %q", n.Pos, n))
			}

		case *DataRef:
			if mdecl := scopes.Resolve(n.Untyped()); mdecl != nil {
				switch decl := mdecl.Decl.(type) {
				case *Data:
					if n.Module == "" {
						n.Module = mdecl.Module.Name
					}
					if len(n.TypeParameters) != len(decl.TypeParameters) {
						merr = append(merr, fmt.Errorf("%s: reference to data structure %s has %d type parameters, but %d were expected",
							n.Pos, n.Name, len(n.TypeParameters), len(decl.TypeParameters)))
					}

				case *TypeParameter:

				default:
					if n.Module == "" {
						merr = append(merr, fmt.Errorf("%s: unqualified reference to invalid data structure %q", n.Pos, n))
					}
					n.Module = mdecl.Module.Name
				}
			} else if n.Module == "" || n.Module == module.Name { // Don't report errors for external modules.
				merr = append(merr, fmt.Errorf("%s: reference to unknown data structure %q", n.Pos, n))
			}

		case *Verb:
			if !ValidateName(n.Name) {
				merr = append(merr, fmt.Errorf("%s: Verb name %q is invalid", n.Pos, n.Name))
			}
			if _, ok := primitivesScope[n.Name]; ok {
				merr = append(merr, fmt.Errorf("%s: Verb name %q is a reserved word", n.Pos, n.Name))
			}

		case *Data:
			if !ValidateName(n.Name) {
				merr = append(merr, fmt.Errorf("%s: data structure name %q is invalid", n.Pos, n.Name))
			}
			if _, ok := primitivesScope[n.Name]; ok {
				merr = append(merr, fmt.Errorf("%s: data structure name %q is a reserved word", n.Pos, n.Name))
			}
			for _, md := range n.Metadata {
				if md, ok := md.(*MetadataCalls); ok {
					merr = append(merr, fmt.Errorf("%s: metadata %q is not valid on data structures", md.Pos, strings.TrimSpace(md.String())))
				}
			}

		case *Array, *Bool, *Database, *Field, *Float, *Int,
			*Time, *Map, *Module, *Schema, *String, *Bytes,
			*MetadataCalls, *MetadataDatabases, *MetadataIngress, IngressPathComponent,
			*IngressPathLiteral, *IngressPathParameter, *Optional,
			*SourceRef, *SinkRef, *Unit, *Any, *TypeParameter:

		case Type, Metadata, Decl: // Union types.
		}
		return next()
	})
	merr = cleanErrors(merr)
	return errors.Join(merr...)
}

// Sort and de-duplicate errors.
func cleanErrors(merr []error) []error {
	if len(merr) == 0 {
		return nil
	}
	// Deduplicate.
	set := map[string]error{}
	for _, err := range merr {
		for _, subErr := range errors.UnwrapAll(err) {
			set[strings.TrimSpace(subErr.Error())] = subErr
		}
	}
	merr = maps.Values(set)
	// Sort by position.
	sort.Slice(merr, func(i, j int) bool {
		var ipe, jpe participle.Error
		if errors.As(merr[i], &ipe) && errors.As(merr[j], &jpe) {
			ipp := ipe.Position()
			jpp := jpe.Position()
			return ipp.Line < jpp.Line || (ipp.Line == jpp.Line && ipp.Column < jpp.Column)
		}
		return merr[i].Error() < merr[j].Error()
	})
	return merr
}
