package schema

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/alecthomas/types/optional"

	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
)

var ErrNotFound = errors.New("not found")

type Schema struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Modules []*Module `parser:"@@*" protobuf:"2"`
}

var _ Node = (*Schema)(nil)

func (s *Schema) Position() Position { return s.Pos }
func (s *Schema) String() string {
	out := &strings.Builder{}
	for i, m := range s.Modules {
		if i != 0 {
			fmt.Fprintln(out)
		}
		fmt.Fprint(out, m)
	}
	return out.String()
}

func (s *Schema) schemaChildren() []Node {
	out := make([]Node, len(s.Modules))
	for i, m := range s.Modules {
		out[i] = m
	}
	return out
}

func (s *Schema) Hash() [sha256.Size]byte {
	return sha256.Sum256([]byte(s.String()))
}

// ResolveRequestResponseType resolves a reference to a supported request/response type, which can be a Data or an Any,
// or a TypeAlias over either supported type.
func (s *Schema) ResolveRequestResponseType(ref *Ref) (Symbol, error) {
	decl, ok := s.Resolve(ref).Get()
	if !ok {
		return nil, fmt.Errorf("unknown ref %s", ref)
	}

	if ta, ok := decl.(*TypeAlias); ok {
		if typ, ok := ta.Type.(*Any); ok {
			return typ, nil
		}
	}

	return s.resolveToSymbolMonomorphised(ref, nil)
}

// ResolveMonomorphised resolves a reference to a monomorphised Data type.
// Also supports resolving the monomorphised Data type underlying a TypeAlias, where applicable.
//
// If a Ref is not found, returns ErrNotFound.
func (s *Schema) ResolveMonomorphised(ref *Ref) (*Data, error) {
	return s.resolveToDataMonomorphised(ref, nil)
}

func (s *Schema) resolveToDataMonomorphised(n Node, parent Node) (*Data, error) {
	switch typ := n.(type) {
	case *Ref:
		resolved, ok := s.Resolve(typ).Get()
		if !ok {
			return nil, fmt.Errorf("unknown ref %s", typ)
		}
		return s.resolveToDataMonomorphised(resolved, typ)
	case *Data:
		p, ok := parent.(*Ref)
		if !ok {
			return nil, fmt.Errorf("expected data node parent to be a ref, got %T", p)
		}
		return typ.Monomorphise(p)
	case *TypeAlias:
		return s.resolveToDataMonomorphised(typ.Type, typ)
	default:
		return nil, fmt.Errorf("expected data or type alias of data, got %T", typ)
	}
}

func (s *Schema) resolveToSymbolMonomorphised(n Node, parent Node) (Symbol, error) {
	switch typ := n.(type) {
	case *Ref:
		resolved, ok := s.Resolve(typ).Get()
		if !ok {
			return nil, fmt.Errorf("unknown ref %s", typ)
		}
		return s.resolveToSymbolMonomorphised(resolved, typ)
	case *Data:
		p, ok := parent.(*Ref)
		if !ok {
			return nil, fmt.Errorf("expected data node parent to be a ref, got %T", p)
		}
		return typ.Monomorphise(p)
	case *TypeAlias:
		return s.resolveToSymbolMonomorphised(typ.Type, typ)
	case Symbol:
		return typ, nil
	default:
		return nil, fmt.Errorf("expected data or type alias of data, got %T", typ)
	}
}

// ResolveWithModule a reference to a declaration and its module.
func (s *Schema) ResolveWithModule(ref *Ref) (optional.Option[Decl], optional.Option[*Module]) {
	for _, module := range s.Modules {
		if module.Name == ref.Module {
			for _, decl := range module.Decls {
				if decl.GetName() == ref.Name {
					return optional.Some(decl), optional.Some(module)
				}
			}
		}
	}
	return optional.None[Decl](), optional.None[*Module]()
}

// Resolve a reference to a declaration.
func (s *Schema) Resolve(ref *Ref) optional.Option[Decl] {
	decl, _ := s.ResolveWithModule(ref)
	return decl
}

// ResolveToType resolves a reference to a declaration of the given type.
//
// The out parameter must be a pointer to a non-nil Decl implementation or this
// will panic.
//
//	data := &Data{}
//	err := s.ResolveToType(ref, data)
func (s *Schema) ResolveToType(ref *Ref, out Decl) error {
	// Programmer error.
	if reflect.ValueOf(out).Kind() != reflect.Ptr {
		panic(fmt.Errorf("out parameter is not a pointer"))
	}
	if reflect.ValueOf(out).Elem().Kind() == reflect.Invalid {
		panic(fmt.Errorf("out parameter is a nil pointer"))
	}

	for _, module := range s.Modules {
		if module.Name == ref.Module {
			for _, decl := range module.Decls {
				if decl.GetName() == ref.Name {
					declType := reflect.TypeOf(decl)
					outType := reflect.TypeOf(out)
					if declType.Elem().AssignableTo(outType.Elem()) {
						reflect.ValueOf(out).Elem().Set(reflect.ValueOf(decl).Elem())
						return nil
					}
					return fmt.Errorf("resolved declaration is not of the expected type: want %s, got %s",
						outType, declType)
				}
			}
		}
	}

	return fmt.Errorf("could not resolve reference %v: %w", ref, ErrNotFound)
}

// Module returns the named module if it exists.
func (s *Schema) Module(name string) optional.Option[*Module] {
	for _, module := range s.Modules {
		if module.Name == name {
			return optional.Some(module)
		}
	}
	return optional.None[*Module]()
}

// Upsert inserts or replaces a module.
func (s *Schema) Upsert(module *Module) {
	for i, m := range s.Modules {
		if m.Name == module.Name {
			s.Modules[i] = module
			return
		}
	}
	s.Modules = append(s.Modules, module)
}

// TypeName returns the name of a type as a string, stripping any package prefix and correctly handling Ref aliases.
func TypeName(v any) string {
	t := reflect.Indirect(reflect.ValueOf(v)).Type()

	// handle AbstractRefs like "AbstractRef[github.com/block/ftl/common/protos/xyz/block/ftl/schema.DataRef]"
	if strings.HasPrefix(t.Name(), "AbstractRef[") {
		return strings.TrimSuffix(strings.Split(t.Name(), ".")[2], "]")
	}

	return t.Name()
}

// FromProto converts a protobuf Schema to a Schema and validates it.
func FromProto(s *schemapb.Schema) (*Schema, error) {
	modules, err := moduleListToSchema(s.Modules)
	if err != nil {
		return nil, err
	}
	schema := &Schema{
		Modules: modules,
	}
	return ValidateSchema(schema)
}

// ModuleDependencies returns the modules that the given module depends on
// Dependency modules are the ones that are called by the given module, or that publish topics that the given module subscribes to
func (s *Schema) ModuleDependencies(module string) map[string]*Module {
	mods := map[string]*Module{}
	for _, sch := range s.Modules {
		mods[sch.Name] = sch
	}
	deps := make(map[string]*Module)
	toProcess := []string{module}
	for len(toProcess) > 0 {
		dep := toProcess[0]
		toProcess = toProcess[1:]
		if deps[dep] != nil {
			continue
		}
		dm := mods[dep]
		deps[dep] = dm
		for _, m := range dm.Decls {
			if ref, ok := m.(*Verb); ok {
				for _, ref := range ref.Metadata {
					switch md := ref.(type) {
					case *MetadataCalls:
						for _, calls := range md.Calls {
							if calls.Module != "" {
								toProcess = append(toProcess, calls.Module)
							}
						}
					case *MetadataSubscriber:
						if md.Topic.Module != "" {
							toProcess = append(toProcess, md.Topic.Module)
						}
					default:
					}
				}
			}
		}
	}
	delete(deps, module)
	return deps
}
