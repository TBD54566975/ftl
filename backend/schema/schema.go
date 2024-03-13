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

func (s *Schema) ResolveDataRefMonomorphised(ref *Ref) (*Data, error) {
	decl := s.ResolveRef(ref)
	if decl == nil {
		return nil, fmt.Errorf("unknown data %v", ref)
	}

	var data *Data
	if d, ok := decl.(*Data); ok {
		data = d
	} else {
		return nil, fmt.Errorf("invalid data reference %v", ref)
	}

	return data.Monomorphise(ref)
}

func (s *Schema) ResolveVerbRef(ref *Ref) *Verb {
	for _, module := range s.Modules {
		if module.Name == ref.Module {
			for _, decl := range module.Decls {
				if verb, ok := decl.(*Verb); ok && verb.Name == ref.Name {
					return verb
				}
			}
		}
	}
	return nil
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

// Module returns the named module if it exists.
func (s *Schema) Module(name string) optional.Option[*Module] {
	for _, module := range s.Modules {
		if module.Name == name {
			return optional.Some(module)
		}
	}
	return optional.None[*Module]()
}

func (s *Schema) DataMap() map[RefKey]*Data {
	dataTypes := map[RefKey]*Data{}
	for _, module := range s.Modules {
		for _, decl := range module.Decls {
			if data, ok := decl.(*Data); ok {
				dataTypes[RefKey{Module: module.Name, Name: data.Name}] = data
			}
		}
	}
	return dataTypes
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
	return Validate(schema)
}
