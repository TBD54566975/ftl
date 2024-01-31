package schema

import (
	"fmt"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

type Module struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" protobuf:"2"`
	Builtin  bool     `parser:"@'builtin'?" protobuf:"3"`
	Name     string   `parser:"'module' @Ident '{'" protobuf:"4"`
	Decls    []Decl   `parser:"@@* '}'" protobuf:"5"`
}

var _ Node = (*Module)(nil)
var _ Decl = (*Module)(nil)

// Scope returns a scope containing all the declarations in this module.
func (m *Module) Scope() Scope {
	scope := Scope{}
	for _, d := range m.Decls {
		switch d := d.(type) {
		case *Data:
			scope[d.Name] = ModuleDecl{m, d}

		case *Verb:
			scope[d.Name] = ModuleDecl{m, d}

		case *Bool, *Bytes, *Database, *Float, *Int, *Module, *String, *Time,
			*Unit, *Any, *TypeParameter:
		}
	}
	return scope
}

// Resolve returns the declaration in this module with the given name, or nil
func (m *Module) Resolve(ref Ref) *ModuleDecl {
	if ref.Module != "" && ref.Module != m.Name {
		return nil
	}
	for _, d := range m.Decls {
		switch d := d.(type) {
		case *Data:
			if d.Name == ref.Name {
				return &ModuleDecl{m, d}
			}
		case *Verb:
			if d.Name == ref.Name {
				return &ModuleDecl{m, d}
			}

		case *Bool, *Bytes, *Database, *Float, *Int, *Module, *String, *Time,
			*Unit, *Any, *TypeParameter:
		}
	}
	return nil
}

func (m *Module) schemaDecl()        {}
func (m *Module) Position() Position { return m.Pos }
func (m *Module) schemaChildren() []Node {
	children := make([]Node, 0, len(m.Decls))
	for _, d := range m.Decls {
		children = append(children, d)
	}
	return children
}
func (m *Module) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, encodeComments(m.Comments))
	if m.Builtin {
		fmt.Fprint(w, "builtin ")
	}
	fmt.Fprintf(w, "module %s {\n", m.Name)
	for i, s := range m.Decls {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintln(w, indent(s.String()))
	}
	fmt.Fprintln(w, "}")
	return w.String()
}

// AddData and return its index.
func (m *Module) AddData(data *Data) int {
	for i, d := range m.Decls {
		if d, ok := d.(*Data); ok && d.Name == data.Name {
			return i
		}
	}
	m.Decls = append(m.Decls, data)
	return len(m.Decls) - 1
}

func (m *Module) Verbs() []*Verb {
	var verbs []*Verb
	for _, d := range m.Decls {
		if v, ok := d.(*Verb); ok {
			verbs = append(verbs, v)
		}
	}
	return verbs
}

func (m *Module) Data() []*Data {
	var data []*Data
	for _, d := range m.Decls {
		if v, ok := d.(*Data); ok {
			data = append(data, v)
		}
	}
	return data
}

// Imports returns the modules imported by this module.
func (m *Module) Imports() []string {
	imports := map[string]bool{}
	_ = Visit(m, func(n Node, next func() error) error {
		switch n := n.(type) {
		case *DataRef:
			if n.Module != "" && n.Module != m.Name {
				imports[n.Module] = true
			}

		case *VerbRef:
			if n.Module != "" && n.Module != m.Name {
				imports[n.Module] = true
			}

		default:
		}
		return next()
	})

	importStrs := maps.Keys(imports)
	sort.Strings(importStrs)
	return importStrs
}

func (m *Module) ToProto() proto.Message {
	return &schemapb.Module{
		Pos:      posToProto(m.Pos),
		Builtin:  m.Builtin,
		Name:     m.Name,
		Comments: m.Comments,
		Decls:    declListToProto(m.Decls),
	}
}

// ModuleFromProto converts a protobuf Module to a Module and validates it.
func ModuleFromProto(s *schemapb.Module) (*Module, error) {
	module := &Module{
		Pos:      posFromProto(s.Pos),
		Builtin:  s.Builtin,
		Name:     s.Name,
		Comments: s.Comments,
		Decls:    declListToSchema(s.Decls),
	}
	return module, ValidateModule(module)
}

func ModuleFromBytes(b []byte) (*Module, error) {
	s := &schemapb.Module{}
	if err := proto.Unmarshal(b, s); err != nil {
		return nil, err
	}
	return ModuleFromProto(s)
}

func ModuleToBytes(m *Module) ([]byte, error) {
	return proto.Marshal(m.ToProto())
}

func moduleListToSchema(s []*schemapb.Module) ([]*Module, error) {
	var out []*Module
	for _, n := range s {
		module, err := ModuleFromProto(n)
		if err != nil {
			return nil, err
		}
		out = append(out, module)
	}
	return out, nil
}
