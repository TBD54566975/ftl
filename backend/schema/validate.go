package schema

import (
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/alecthomas/participle/v2"
	xreflect "golang.design/x/reflect"
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
	schema = xreflect.DeepCopy(schema)
	modules := map[string]bool{}
	merr := []error{}
	ingress := map[string]*Verb{}

	// Inject builtins.
	builtins := Builtins()
	// Move builtins to the front of the list.
	schema.Modules = slices.DeleteFunc(schema.Modules, func(m *Module) bool { return m.Name == builtins.Name })
	schema.Modules = append([]*Module{builtins}, schema.Modules...)

	scopes := NewScopes()

	// Validate dependencies
	if err := validateDependencies(schema); err != nil {
		merr = append(merr, err)
	}

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
			case *Ref:
				if mdecl := scopes.Resolve(*n); mdecl != nil {
					switch decl := mdecl.Decl.(type) {
					case *Verb, *Enum:
						if mdecl.Module != nil {
							n.Module = mdecl.Module.Name
						}
						if len(n.TypeParameters) != 0 {
							merr = append(merr, fmt.Errorf("%s: reference to %s %q cannot have type parameters",
								n.Pos, reflect.TypeOf(decl).Elem().Name(), n.Name))
						}
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
						merr = append(merr, fmt.Errorf("%s: invalid reference %q at %q", n.Pos, n, mdecl.Decl.Position()))

					}
				} else {
					merr = append(merr, fmt.Errorf("%s: unknown reference %q", n.Pos, n))
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

			case *Enum:
				switch t := n.Type.(type) {
				case *String, *Int:
					for _, v := range n.Variants {
						if reflect.TypeOf(v.Value.schemaValueType()) != reflect.TypeOf(t) {
							merr = append(merr, fmt.Errorf("%s: enum variant %q of type %s cannot have a "+
								"value of type %s", v.Pos, v.Name, t, v.Value.schemaValueType()))
						}
					}
					return next()
				default:
					merr = append(merr, fmt.Errorf("%s: enum type must be String or Int, not %s", n.Pos, n.Type))
				}

			case *Array, *Bool, *Bytes, *Data, *Database, Decl, *Field, *Float,
				IngressPathComponent, *IngressPathLiteral, *IngressPathParameter,
				*Int, *Map, Metadata, *MetadataCalls, *MetadataDatabases,
				*MetadataIngress, *MetadataAlias, *Module, *Optional, *Schema,
				*String, *Time, Type, *Unit, *Any, *TypeParameter, *EnumVariant,
				Value, *IntValue, *StringValue, *Config, *Secret:
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
		case *Ref:
			if mdecl := scopes.Resolve(*n); mdecl != nil {
				switch decl := mdecl.Decl.(type) {
				case *Verb, *Enum:
					n.Module = mdecl.Module.Name
					if len(n.TypeParameters) != 0 {
						merr = append(merr, fmt.Errorf("%s: reference to %s %q cannot have type parameters",
							n.Pos, reflect.TypeOf(decl).Elem().Name(), n.Name))
					}
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
						merr = append(merr, fmt.Errorf("%s: unqualified reference to invalid %s %q", n.Pos, reflect.TypeOf(decl).Elem().Name(), n))
					}
					n.Module = mdecl.Module.Name
				}
			} else if n.Module == "" || n.Module == module.Name { // Don't report errors for external modules.
				merr = append(merr, fmt.Errorf("%s: unknown reference %q", n.Pos, n))
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

		case *Config:
			if !ValidateName(n.Name) {
				merr = append(merr, fmt.Errorf("%s: config name %q is invalid", n.Pos, n.Name))
			}
			if _, ok := primitivesScope[n.Name]; ok {
				merr = append(merr, fmt.Errorf("%s: config name %q is a reserved word", n.Pos, n.Name))
			}

		case *Secret:
			if !ValidateName(n.Name) {
				merr = append(merr, fmt.Errorf("%s: secret name %q is invalid", n.Pos, n.Name))
			}
			if _, ok := primitivesScope[n.Name]; ok {
				merr = append(merr, fmt.Errorf("%s: secret name %q is a reserved word", n.Pos, n.Name))
			}

		case *Field:
			for _, md := range n.Metadata {
				if _, ok := md.(*MetadataAlias); !ok {
					merr = append(merr, fmt.Errorf("%s: metadata %q is not valid on fields", md.Position(), strings.TrimSpace(md.String())))
				}
			}

		case *Array, *Bool, *Database, *Float, *Int,
			*Time, *Map, *Module, *Schema, *String, *Bytes,
			*MetadataCalls, *MetadataDatabases, *MetadataIngress, *MetadataAlias,
			IngressPathComponent, *IngressPathLiteral, *IngressPathParameter, *Optional,
			*Unit, *Any, *TypeParameter, *Enum, *EnumVariant, *IntValue, *StringValue:

		case Type, Metadata, Decl, Value: // Union types.
		}
		return next()
	})
	merr = cleanErrors(merr)
	sort.SliceStable(module.Decls, func(i, j int) bool {
		iDecl := module.Decls[i]
		jDecl := module.Decls[j]
		iType := reflect.TypeOf(iDecl).String()
		jType := reflect.TypeOf(jDecl).String()
		if iType == jType {
			return iDecl.GetName() < jDecl.GetName()
		}
		return iType < jType
	})
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

type dependencyVertex struct {
	from, to string
}

type dependencyVertexState int

const (
	notExplored dependencyVertexState = iota
	exploring
	fullyExplored
)

func validateDependencies(schema *Schema) error {
	// go through schema's modules, find cycles in modules' dependencies

	// First pass, set up direct imports and vertex states for each module
	// We need each import array and vertex array to be sorted to make the output deterministic
	imports := map[string][]string{}
	vertexes := []dependencyVertex{}
	vertexStates := map[dependencyVertex]dependencyVertexState{}

	for _, module := range schema.Modules {
		currentImports := module.Imports()
		sort.Strings(currentImports)
		imports[module.Name] = currentImports

		for _, imp := range currentImports {
			v := dependencyVertex{module.Name, imp}
			vertexes = append(vertexes, v)
			vertexStates[v] = notExplored
		}
	}

	sort.Slice(vertexes, func(i, j int) bool {
		lhs := vertexes[i]
		rhs := vertexes[j]
		return lhs.from < rhs.from || (lhs.from == rhs.from && lhs.to < rhs.to)
	})

	// DFS to find cycles
	for _, v := range vertexes {
		if cycle := dfsForDependencyCycle(imports, vertexStates, v); cycle != nil {
			return fmt.Errorf("found cycle in dependencies: %s", strings.Join(cycle, " -> "))
		}
	}

	return nil
}

func dfsForDependencyCycle(imports map[string][]string, vertexStates map[dependencyVertex]dependencyVertexState, v dependencyVertex) []string {
	switch vertexStates[v] {
	case notExplored:
		vertexStates[v] = exploring

		for _, toModule := range imports[v.to] {
			nextV := dependencyVertex{v.to, toModule}
			if cycle := dfsForDependencyCycle(imports, vertexStates, nextV); cycle != nil {
				// found cycle. prepend current module to cycle and return
				cycle = append([]string{nextV.from}, cycle...)
				return cycle
			}
		}
		vertexStates[v] = fullyExplored
		return nil
	case exploring:
		return []string{v.to}
	case fullyExplored:
		return nil
	}

	return nil
}
