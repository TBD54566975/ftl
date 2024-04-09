package schema

import (
	"crypto/sha256"
	"fmt"
	"reflect"
	"strings"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

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

func (s *Schema) ResolveRefMonomorphised(ref *Ref) (*Data, error) {
	out := &Data{}
	err := s.ResolveRefToType(ref, out)
	if err != nil {
		return nil, err
	}
	return out.Monomorphise(ref)
}

func (s *Schema) ResolveRef(ref *Ref) Decl {
	for _, module := range s.Modules {
		if module.Name == ref.Module {
			for _, decl := range module.Decls {
				if decl.GetName() == ref.Name {
					return decl
				}
			}
		}
	}
	return nil
}

func (s *Schema) ResolveRefToType(ref *Ref, out Decl) error {
	if reflect.ValueOf(out).Kind() != reflect.Ptr {
		return fmt.Errorf("out parameter is not a pointer")
	}
	if reflect.ValueOf(out).Elem().Kind() == reflect.Invalid {
		return fmt.Errorf("out parameter is a nil pointer")
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
	return fmt.Errorf("could not resolve reference %v", ref)
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

func (s *Schema) ToProto() proto.Message {
	return &schemapb.Schema{
		Pos:     posToProto(s.Pos),
		Modules: nodeListToProto[*schemapb.Module](s.Modules),
	}
}

func TypeName(v any) string {
	t := reflect.Indirect(reflect.ValueOf(v)).Type()

	// handle AbstractRefs like "AbstractRef[github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema.DataRef]"
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
