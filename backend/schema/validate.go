package schema

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"golang.design/x/reflect"
)

var (
	// Identifiers that can't be used as data or verb names.
	//
	// We don't need Array/Map/VerbRef/DataRef here because there are no
	// keywords associated with these types.
	reservedIdentNames = map[string]bool{
		"Int": true, "Float": true, "String": true, "Bytes": true, "Bool": true,
		"Time": true, "Unit": true,
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
	verbs := map[string]bool{}
	data := map[string]bool{}
	merr := []error{}
	ingress := map[string]*Verb{}
	verbRefs := []*VerbRef{}
	dataRefs := []*DataRef{}

	// Inject builtins.
	builtins := Builtins()
	schema.Modules = slices.DeleteFunc(schema.Modules, func(m *Module) bool { return m.Name == builtins.Name })
	schema.Modules = append([]*Module{builtins}, schema.Modules...)

	// Validate modules.
	for _, module := range schema.Modules {
		if _, seen := modules[module.Name]; seen {
			merr = append(merr, fmt.Errorf("%s: duplicate module %q", module.Pos, module.Name))
		}
		modules[module.Name] = true
		if err := ValidateModule(module); err != nil {
			merr = append(merr, err)
		}
		// Note that we don't need to check ref names here because the targets
		// themselves must be valid, and the refs cannot refer to non-existent
		// targets.
		err := Visit(module, func(n Node, next func() error) error {
			switch n := n.(type) {
			case *VerbRef:
				verbRefs = append(verbRefs, n)

			case *DataRef:
				dataRefs = append(dataRefs, n)

			case *Verb:
				for _, md := range n.Metadata {
					if md, ok := md.(*MetadataIngress); ok {
						if existing, ok := ingress[md.String()]; ok {
							return fmt.Errorf("duplicate %q for %s:%q and %s:%q", md.String(), existing.Pos, existing.Name, n.Pos, n.Name)
						}
						if md.Type == "http" && (n.Request.String() != "builtin.HttpRequest" || n.Response.String() != "builtin.HttpResponse") {
							return fmt.Errorf("%s: HTTP ingress verb %s(%s) %s must have the signature %s(builtin.HttpRequest) builtin.HttpResponse",
								md.Pos, n.Name, n.Request, n.Response, n.Name)
						}
						ingress[md.String()] = n
					}
				}
				ref := makeRef(module.Name, n.Name)
				verbs[ref] = true
				verbs[n.Name] = true

			case *Data:
				ref := makeRef(module.Name, n.Name)
				data[ref] = true
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
		if ref.Module != "" && !verbs[ref.String()] {
			merr = append(merr, fmt.Errorf("%s: reference to unknown Verb %q", ref.Pos, ref))
		}
	}
	for _, ref := range dataRefs {
		if ref.Module != "" && !data[ref.String()] {
			merr = append(merr, fmt.Errorf("%s: reference to unknown data structure %q", ref.Pos, ref))
		}
	}
	return schema, errors.Join(merr...)
}

var validNameRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ValidateModule performs the subset of semantic validation possible on a single module.
func ValidateModule(module *Module) error {
	verbs := map[string]bool{}
	data := map[string]bool{}
	verbRefs := []*VerbRef{}
	dataRefs := []*DataRef{}
	merr := []error{}

	if !validNameRe.MatchString(module.Name) {
		merr = append(merr, fmt.Errorf("%s: module name %q is invalid", module.Pos, module.Name))
	}
	if module.Builtin && module.Name != "builtin" {
		merr = append(merr, fmt.Errorf("%s: only the \"ftl\" module can be marked as builtin", module.Pos))
	}
	visit := func(module *Module) error {
		return Visit(module, func(n Node, next func() error) error {
			switch n := n.(type) {
			case *VerbRef:
				verbRefs = append(verbRefs, n)

			case *DataRef:
				dataRefs = append(dataRefs, n)

			case *Verb:
				if !validNameRe.MatchString(n.Name) {
					merr = append(merr, fmt.Errorf("%s: Verb name %q is invalid", n.Pos, n.Name))
				}
				if _, ok := reservedIdentNames[n.Name]; ok {
					merr = append(merr, fmt.Errorf("%s: Verb name %q is a reserved word", n.Pos, n.Name))
				}
				if _, ok := verbs[n.Name]; ok {
					merr = append(merr, fmt.Errorf("%s: duplicate Verb %q", n.Pos, n.Name))
				}
				verbs[module.Name+"."+n.Name] = true

			case *Data:
				if !validNameRe.MatchString(n.Name) {
					merr = append(merr, fmt.Errorf("%s: data structure name %q is invalid", n.Pos, n.Name))
				}
				if _, ok := reservedIdentNames[n.Name]; ok {
					merr = append(merr, fmt.Errorf("%s: data structure name %q is a reserved word", n.Pos, n.Name))
				}
				if _, ok := data[n.Name]; ok {
					merr = append(merr, fmt.Errorf("%s: duplicate data structure %q", n.Pos, n.Name))
				}
				for _, md := range n.Metadata {
					if md, ok := md.(*MetadataCalls); ok {
						merr = append(merr, fmt.Errorf("%s: metadata %q is not valid on data structures", md.Pos, strings.TrimSpace(md.String())))
					}
				}
				data[module.Name+"."+n.Name] = true

			case *Array, *Bool, *Database, *Field, *Float, *Int,
				*Time, *Map, *Module, *Schema, *String, *Bytes,
				*MetadataCalls, *MetadataDatabases, *MetadataIngress, IngressPathComponent,
				*IngressPathLiteral, *IngressPathParameter, *Optional,
				*SourceRef, *SinkRef, *Unit:

			case Type, Metadata, Decl: // Union types.
			}
			return next()
		})
	}

	// Collect all the builtin verbs and data structures so that unqualified refs in our module can be resolved.
	builtins := Builtins()
	if err := visit(builtins); err != nil {
		return err
	}
	// Clear collected refs because we don't care about any that might be in the builtins.
	verbRefs = nil
	dataRefs = nil

	// Collect verbs and data structures from this module.
	if err := visit(module); err != nil {
		return err
	}

	for _, ref := range verbRefs {
		if !resolveLocalRef(module, (*Ref)(ref), verbs) {
			merr = append(merr, fmt.Errorf("%s: unqualified reference to unknown Verb %q", ref.Pos, ref))
		}
	}
	for _, ref := range dataRefs {
		if !resolveLocalRef(module, (*Ref)(ref), data) {
			merr = append(merr, fmt.Errorf("%s: unqualified reference to unknown data structure %q", ref.Pos, ref))
		}
	}
	return errors.Join(merr...)
}

// Try to resolve a relative reference (ie. one without a module).
//
// This first tries to resolve the reference against the local module, then against
// the builtins module.
func resolveLocalRef(localModule *Module, ref *Ref, exist map[string]bool) bool {
	if ref.Module != "" {
		return true
	}
	for _, module := range []string{localModule.Name, "builtin"} {
		clone := reflect.DeepCopy(ref)
		clone.Module = module
		if exist[clone.String()] {
			ref.Module = module
			return true
		}
	}
	return false
}
