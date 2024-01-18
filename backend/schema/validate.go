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
	}
)

type Resolver interface {
	// Resolve a reference to a symbol declaration or nil.
	Resolve(ref Ref) *ModuleDecl
}

// Scope maps relative names to fully qualified types.
type Scope map[string]ModuleDecl

// ModuleDecl is a declaration associated with a module.
type ModuleDecl struct {
	Module *Module // May be nil.
	Decl   Decl
}

func (s Scope) String() string {
	out := &strings.Builder{}
	for name, decl := range s {
		fmt.Fprintf(out, "%s: %T\n", name, decl.Decl)
	}
	return out.String()
}

// Scopes to search during type resolution.
type Scopes []Scope

var _ Resolver = Scopes{}

func (s Scopes) String() string {
	out := &strings.Builder{}
	for i, scope := range s {
		if i != 0 {
			fmt.Fprintln(out)
		}
		fmt.Fprintf(out, "Scope %d:\n", i)
		for name, decl := range scope {
			fmt.Fprintf(out, "  %s: %T\n", name, decl.Decl)
		}
	}
	return out.String()
}

// Push a new Scope onto the stack.
//
// This contains references to previous Scopes so that updates are preserved.
func (s Scopes) Push() Scopes {
	out := make(Scopes, 0, len(s)+1)
	out = append(out, s...)
	out = append(out, Scope{})
	return out
}

// PushModule pushes a new Scope onto the stack containing all the declarations
// in the module.
//
// The module itself is added to the current scope before pushing.
func (s *Scopes) PushModule(module *Module) (Scopes, error) {
	if err := s.Add(nil, module.Name, module); err != nil {
		return nil, err
	}
	out := s.Push()
	for _, decl := range module.Decls {
		switch decl := decl.(type) {
		case *Data:
			if err := out.Add(module, decl.Name, decl); err != nil {
				return nil, err
			}

		case *Verb:
			if err := out.Add(module, decl.Name, decl); err != nil {
				return nil, err
			}

		case *Bool, *Bytes, *Database, *Float, *Int, *Module, *String, *Time, *Unit:
		}
	}
	return out, nil
}

// Add a declaration to the current scope.
func (s *Scopes) Add(owner *Module, name string, decl Decl) error {
	end := len(*s) - 1
	if prev, ok := (*s)[end][name]; ok {
		return fmt.Errorf("%s: duplicate declaration of %q at %s", decl.Position(), name, prev.Decl.Position())
	}
	(*s)[end][name] = ModuleDecl{owner, decl}
	return nil
}

// Resolve a reference to a symbol declaration or nil.
func (s Scopes) Resolve(ref Ref) *ModuleDecl {
	if ref.Module == "" {
		for i := len(s) - 1; i >= 0; i-- {
			scope := s[i]
			if decl, ok := scope[ref.Name]; ok {
				return &decl
			}
		}
		return nil
	}
	// If a module is provided, try to resolve it, then resolve the reference through the module.
	for i := len(s) - 1; i >= 0; i-- {
		scope := s[i]
		if mdecl, ok := scope[ref.Module]; ok {
			if resolver, ok := mdecl.Decl.(Resolver); ok {
				if decl := resolver.Resolve(ref); decl != nil {
					// Holy nested if statement Batman.
					return decl
				}
			}
		}
	}
	return nil
}

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

	// Add types from the builtins module to the scope stack.
	scopes := Scopes{primitivesScope, Scope{}}
	var err error
	scopes, err = scopes.PushModule(builtins)
	if err != nil {
		return nil, err
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

		moduleScopes, err := scopes.PushModule(module)
		if err != nil {
			merr = append(merr, err)
		}

		err = Visit(module, func(n Node, next func() error) error {
			switch n := n.(type) {
			case *VerbRef:
				if mdecl := moduleScopes.Resolve(n.Untyped()); mdecl != nil {
					if _, ok := mdecl.Decl.(*Verb); !ok {
						merr = append(merr, fmt.Errorf("%s: reference to invalid verb %q at %q", n.Pos, n, mdecl.Decl.Position()))
					} else if mdecl.Module != nil {
						n.Module = mdecl.Module.Name
					}
				} else {
					merr = append(merr, fmt.Errorf("%s: reference to unknown verb %q", n.Pos, n))
				}

			case *DataRef:
				if mdecl := moduleScopes.Resolve(n.Untyped()); mdecl != nil {
					if _, ok := mdecl.Decl.(*Data); !ok {
						merr = append(merr, fmt.Errorf("%s: reference to invalid data structure %q at %s", n.Pos, n, mdecl.Decl.Position()))
					} else if mdecl.Module == nil {
						n.Module = mdecl.Module.Name
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
						if md.Type == "http" && (n.Request.String() != "builtin.HttpRequest" || n.Response.String() != "builtin.HttpResponse") {
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
				*Unit:
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
func ValidateModule(module *Module) error {
	verbRefs := []*VerbRef{}
	dataRefs := []*DataRef{}
	merr := []error{}

	scopes := Scopes{primitivesScope, Scope{}}
	scopes, err := scopes.PushModule(Builtins())
	if err != nil {
		return err
	}

	if !ValidateName(module.Name) {
		merr = append(merr, fmt.Errorf("%s: module name %q is invalid", module.Pos, module.Name))
	}
	if module.Builtin && module.Name != "builtin" {
		merr = append(merr, fmt.Errorf("%s: only the \"ftl\" module can be marked as builtin", module.Pos))
	}
	if err := scopes.Add(nil, module.Name, module); err != nil {
		merr = append(merr, err)
	}
	scopes = scopes.Push() //nolint:govet
	_ = Visit(module, func(n Node, next func() error) error {
		switch n := n.(type) {
		case *VerbRef:
			verbRefs = append(verbRefs, n)

		case *DataRef:
			dataRefs = append(dataRefs, n)

		case *Verb:
			if !ValidateName(n.Name) {
				merr = append(merr, fmt.Errorf("%s: Verb name %q is invalid", n.Pos, n.Name))
			}
			if _, ok := primitivesScope[n.Name]; ok {
				merr = append(merr, fmt.Errorf("%s: Verb name %q is a reserved word", n.Pos, n.Name))
			}
			if err := scopes.Add(module, n.Name, n); err != nil {
				merr = append(merr, err)
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
			if err := scopes.Add(module, n.Name, n); err != nil {
				merr = append(merr, err)
			}

		case *Array, *Bool, *Database, *Field, *Float, *Int,
			*Time, *Map, *Module, *Schema, *String, *Bytes,
			*MetadataCalls, *MetadataDatabases, *MetadataIngress, IngressPathComponent,
			*IngressPathLiteral, *IngressPathParameter, *Optional,
			*SourceRef, *SinkRef, *Unit:

		case Type, Metadata, Decl: // Union types.
		}
		return next()
	})

	for _, ref := range verbRefs {
		if mdecl := scopes.Resolve(ref.Untyped()); mdecl != nil {
			if _, ok := mdecl.Decl.(*Verb); !ok && ref.Module == "" {
				merr = append(merr, fmt.Errorf("%s: unqualified reference to invalid verb %q", ref.Pos, ref))
			} else {
				ref.Module = mdecl.Module.Name
			}
		} else if ref.Module == "" || ref.Module == module.Name { // Don't report errors for external modules.
			merr = append(merr, fmt.Errorf("%s: reference to unknown verb %q", ref.Pos, ref))
		}
	}
	for _, ref := range dataRefs {
		if mdecl := scopes.Resolve(ref.Untyped()); mdecl != nil {
			if _, ok := mdecl.Decl.(*Data); !ok && ref.Module == "" {
				merr = append(merr, fmt.Errorf("%s: unqualified reference to invalid data structure %q", ref.Pos, ref))
			} else {
				ref.Module = mdecl.Module.Name
			}
		} else if ref.Module == "" || ref.Module == module.Name { // Don't report errors for external modules.
			merr = append(merr, fmt.Errorf("%s: reference to unknown data structure %q", ref.Pos, ref))
		}
	}
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
