//nolint:nakedret
package schema

import (
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/alecthomas/types/optional"
	xreflect "golang.design/x/reflect"
	"golang.org/x/exp/maps"

	"github.com/TBD54566975/ftl/internal/errors"
)

var (
	// Primitive types can't be used as identifiers.
	//
	// We don't need Array/Map/Ref here because there are no
	// keywords associated with these types.
	primitivesScope = Scope{
		"Int":    ModuleDecl{Symbol: &Int{}},
		"Float":  ModuleDecl{Symbol: &Float{}},
		"String": ModuleDecl{Symbol: &String{}},
		"Bytes":  ModuleDecl{Symbol: &Bytes{}},
		"Bool":   ModuleDecl{Symbol: &Bool{}},
		"Time":   ModuleDecl{Symbol: &Time{}},
		"Unit":   ModuleDecl{Symbol: &Unit{}},
		"Any":    ModuleDecl{Symbol: &Any{}},
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
		if err := scopes.Add(optional.None[*Module](), module.Name, module); err != nil {
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
			merr = append(merr, errorf(module, "duplicate module %q", module.Name))
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
					switch decl := mdecl.Symbol.(type) {
					case *Verb, *Enum, *Database, *Config, *Secret:
						if module, ok := mdecl.Module.Get(); ok {
							n.Module = module.Name
						}
						if len(n.TypeParameters) != 0 {
							merr = append(merr, errorf(n, "reference to %s %q cannot have type parameters", typeName(decl), n.Name))
						}
					case *Data:
						if module, ok := mdecl.Module.Get(); ok {
							n.Module = module.Name
						}
						if len(n.TypeParameters) != len(decl.TypeParameters) {
							merr = append(merr, errorf(n, "reference to data structure %s has %d type parameters, but %d were expected",
								n.Name, len(n.TypeParameters), len(decl.TypeParameters)))
						}

					case *TypeParameter:
					default:
						merr = append(merr, errorf(n, "invalid reference %q at %q", n, mdecl.Symbol.Position()))
					}
				} else {
					merr = append(merr, errorf(n, "unknown reference %q", n))
				}

			case *Verb:
				for _, md := range n.Metadata {
					if md, ok := md.(*MetadataIngress); ok {
						// Check for duplicate ingress keys
						key := md.Method + " "
						for _, path := range md.Path {
							switch path := path.(type) {
							case *IngressPathLiteral:
								key += "/" + path.Text

							case *IngressPathParameter:
								key += "/{}"
							}
						}
						if existing, ok := ingress[key]; ok {
							merr = append(merr, errorf(md, "duplicate %s ingress %s for %s:%q and %s:%q", md.Type, key, existing.Pos, existing.Name, n.Pos, n.Name))
						}
						ingress[key] = n
					}
				}

			case *Enum:
				switch t := n.Type.(type) {
				case *String, *Int:
					for _, v := range n.Variants {
						if reflect.TypeOf(v.Value.schemaValueType()) != reflect.TypeOf(t) {
							merr = append(merr, errorf(v, "enum variant %q of type %s cannot have a value of type %s", v.Name, t, v.Value.schemaValueType()))
						}
					}
					return next()
				default:
					merr = append(merr, errorf(n, "enum type must be String or Int, not %s", n.Type))
				}

			case *Array, *Bool, *Bytes, *Data, *Database, Decl, *Field, *Float,
				IngressPathComponent, *IngressPathLiteral, *IngressPathParameter,
				*Int, *Map, Metadata, *MetadataCalls, *MetadataDatabases,
				*MetadataIngress, *MetadataAlias, *Module, *Optional, *Schema,
				*String, *Time, Type, *Unit, *Any, *TypeParameter, *EnumVariant,
				Value, *IntValue, *StringValue, *Config, *Secret, Symbol, Named:
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
		merr = append(merr, errorf(module, "module name %q is invalid", module.Name))
	}
	if module.Builtin && module.Name != "builtin" {
		merr = append(merr, errorf(module, "the \"builtin\" module is reserved"))
	}
	if err := scopes.Add(optional.None[*Module](), module.Name, module); err != nil {
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
			if scopes.Resolve(*n) == nil && n.Module == "" {
				merr = append(merr, errorf(n, "unknown reference %q", n))
			}
			if mdecl := scopes.Resolve(*n); mdecl != nil {
				moduleName := ""
				if m, ok := mdecl.Module.Get(); ok {
					moduleName = m.Name
				}
				switch decl := mdecl.Symbol.(type) {
				case *Verb, *Enum, *Database, *Config, *Secret:
					if n.Module == "" {
						n.Module = moduleName
					}
					if len(n.TypeParameters) != 0 {
						merr = append(merr, errorf(n, "reference to %s %q cannot have type parameters", typeName(decl), n.Name))
					}
				case *Data:
					if n.Module == "" {
						n.Module = moduleName
					}
					if len(n.TypeParameters) != len(decl.TypeParameters) {
						merr = append(merr, errorf(n, "reference to data structure %s has %d type parameters, but %d were expected",
							n.Name, len(n.TypeParameters), len(decl.TypeParameters)))
					}
				case *TypeParameter:
				default:
					if n.Module == "" {
						merr = append(merr, errorf(n, "unqualified reference to invalid %s %q", typeName(decl), n))
					}
				}
			} else if n.Module == "" || n.Module == module.Name { // Don't report errors for external modules.
				merr = append(merr, errorf(n, "unknown reference %q", n))
			}

		case *Verb:
			if !ValidateName(n.Name) {
				merr = append(merr, errorf(n, "Verb name %q is invalid", n.Name))
			}
			if _, ok := primitivesScope[n.Name]; ok {
				merr = append(merr, errorf(n, "Verb name %q is a reserved word", n.Name))
			}

			merr = append(merr, validateVerbMetadata(scopes, n)...)

		case *Data:
			if !ValidateName(n.Name) {
				merr = append(merr, errorf(n, "data structure name %q is invalid", n.Name))
			}
			if _, ok := primitivesScope[n.Name]; ok {
				merr = append(merr, errorf(n, "data structure name %q is a reserved word", n.Name))
			}
			for _, md := range n.Metadata {
				if md, ok := md.(*MetadataCalls); ok {
					merr = append(merr, errorf(md, "metadata %q is not valid on data structures", strings.TrimSpace(md.String())))
				}
			}

		case *Config:
			if !ValidateName(n.Name) {
				merr = append(merr, errorf(n, "config name %q is invalid", n.Name))
			}
			if _, ok := primitivesScope[n.Name]; ok {
				merr = append(merr, errorf(n, "config name %q is a reserved word", n.Name))
			}

		case *Secret:
			if !ValidateName(n.Name) {
				merr = append(merr, errorf(n, "secret name %q is invalid", n.Name))
			}
			if _, ok := primitivesScope[n.Name]; ok {
				merr = append(merr, errorf(n, "secret name %q is a reserved word", n.Name))
			}

		case *Field:
			for _, md := range n.Metadata {
				if _, ok := md.(*MetadataAlias); !ok {
					merr = append(merr, errorf(md, "metadata %q is not valid on fields", strings.TrimSpace(md.String())))
				}
			}

		case *Array, *Bool, *Database, *Float, *Int,
			*Time, *Map, *Module, *Schema, *String, *Bytes,
			*MetadataCalls, *MetadataDatabases, *MetadataIngress, *MetadataAlias,
			IngressPathComponent, *IngressPathLiteral, *IngressPathParameter, *Optional,
			*Unit, *Any, *TypeParameter, *Enum, *EnumVariant, *IntValue, *StringValue:

		case Named, Symbol, Type, Metadata, Decl, Value: // Union types.
		}
		return next()
	})
	merr = cleanErrors(merr)
	sort.SliceStable(module.Decls, func(i, j int) bool {
		iDecl := module.Decls[i]
		jDecl := module.Decls[j]
		iPriority := getDeclSortingPriority(iDecl)
		jPriority := getDeclSortingPriority(jDecl)
		if iPriority == jPriority {
			return iDecl.GetName() < jDecl.GetName()
		}
		return iPriority < jPriority
	})
	return errors.Join(merr...)
}

// getDeclSortingPriority (used for schema sorting) is pulled out into it's own switch so the Go sumtype check will fail
// if a new Decl is added but not explicitly prioritized
func getDeclSortingPriority(decl Decl) int {
	priority := 0
	switch decl.(type) {
	case *Config:
		priority = 1
	case *Secret:
		priority = 2
	case *Database:
		priority = 3
	case *Enum:
		priority = 4
	case *Data:
		priority = 5
	case *Verb:
		priority = 6
	}
	return priority
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

func errorf(pos interface{ Position() Position }, format string, args ...interface{}) error {
	return participle.Errorf(lexer.Position(pos.Position()), format, args...)
}

func validateVerbMetadata(scopes Scopes, n *Verb) (merr []error) {
	// Validate metadata
	for _, md := range n.Metadata {
		switch md := md.(type) {
		case *MetadataIngress:
			reqBodyType, reqBody, errs := validateIngressRequestOrResponse(scopes, n, "request", n.Request)
			merr = append(merr, errs...)
			_, _, errs = validateIngressRequestOrResponse(scopes, n, "response", n.Response)
			merr = append(merr, errs...)

			// Validate path
			for _, path := range md.Path {
				switch path := path.(type) {
				case *IngressPathParameter:
					reqBodyData, ok := reqBody.(*Data)
					if !ok {
						merr = append(merr, errorf(path, "ingress verb %s: cannot use path parameter %q with request type %s, expected Data type", n.Name, path.Name, reqBodyType))
					} else if reqBodyData.FieldByName(path.Name) == nil {
						merr = append(merr, errorf(path, "ingress verb %s: request type %s does not contain a field corresponding to the parameter %q", n.Name, reqBodyType, path.Name))
					}

				case *IngressPathLiteral:
				}
			}

		case *MetadataCalls, *MetadataDatabases, *MetadataAlias:
		}
	}
	return
}

func validateIngressRequestOrResponse(scopes Scopes, n *Verb, reqOrResp string, r Type) (fieldType Type, body Symbol, merr []error) {
	rref, _ := r.(*Ref)
	resp, sym := ResolveTypeAs[*Data](scopes, r)
	module, _ := sym.Module.Get()
	if sym == nil || module == nil || module.Name != "builtin" || resp.Name != "Http"+strings.Title(reqOrResp) {
		merr = append(merr, errorf(r, "ingress verb %s: %s type %s must be builtin.HttpRequest", n.Name, reqOrResp, r))
		return
	}

	resp, err := resp.Monomorphise(rref) //nolint:govet
	if err != nil {
		merr = append(merr, errorf(r, "ingress verb %s: %s type %s: %v", n.Name, reqOrResp, r, err))
		return
	}

	scopes = scopes.PushScope(resp.Scope())
	fieldType = resp.FieldByName("body").Type
	if opt, ok := fieldType.(*Optional); ok {
		fieldType = opt.Type
	}
	bodySym := scopes.ResolveType(fieldType)
	if bodySym == nil {
		merr = append(merr, errorf(resp, "ingress verb %s: couldn't resolve %s body type %s", n.Name, reqOrResp, fieldType))
		return
	}
	body = bodySym.Symbol
	switch bodySym.Symbol.(type) {
	case *Bytes, *String, *Data, *Unit, *Float, *Int, *Bool, *Map, *Array: // Valid HTTP response payload types.
	default:
		merr = append(merr, errorf(r, "ingress verb %s: %s type %s must have a body of bytes, string, data structure, unit, float, int, bool, map, or array not %s", n.Name, reqOrResp, r, bodySym.Symbol))
	}
	return
}

// Give a type a human-readable name.
func typeName(v any) string {
	return reflect.Indirect(reflect.ValueOf(v)).Type().Name()
}
