package schema

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"reflect"
	"slices"
	"sort"
	"strings"

	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type Module struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" protobuf:"2"`
	Builtin  bool     `parser:"@'builtin'?" protobuf:"3"`
	Name     string   `parser:"'module' @Ident '{'" protobuf:"4"`
	Decls    []Decl   `parser:"@@* '}'" protobuf:"5"`
}

var _ Node = (*Module)(nil)
var _ Symbol = (*Module)(nil)
var _ sql.Scanner = (*Module)(nil)
var _ driver.Valuer = (*Module)(nil)

func (m *Module) Value() (driver.Value, error) { return proto.Marshal(m.ToProto()) }
func (m *Module) Scan(src any) error {
	switch src := src.(type) {
	case []byte:
		module, err := ModuleFromBytes(src)
		if err != nil {
			return err
		}
		*m = *module
		return nil
	default:
		return fmt.Errorf("cannot scan %T", src)
	}
}

// Resolve returns the declaration in this module with the given name, or nil
func (m *Module) Resolve(ref Ref) *ModuleDecl {
	if ref.Module != "" && ref.Module != m.Name {
		return nil
	}
	for _, d := range m.Decls {
		if d.GetName() == ref.Name {
			return &ModuleDecl{optional.Some(m), d}
		}
	}
	return nil
}

func (m *Module) schemaSymbol()      {}
func (m *Module) Position() Position { return m.Pos }
func (m *Module) schemaChildren() []Node {
	children := make([]Node, 0, len(m.Decls))
	for _, d := range m.Decls {
		children = append(children, d)
	}
	return children
}

type spacingRule struct {
	gapWithinType     bool
	skipGapAfterTypes []reflect.Type
}

func (m *Module) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(m.Comments))
	if m.Builtin {
		fmt.Fprint(w, "builtin ")
	}
	fmt.Fprintf(w, "module %s {\n", m.Name)

	// Print decls with spacing rules
	// Keep these in sync with frontend/console/src/features/modules/schema/schema.utils.ts
	typeSpacingRules := map[reflect.Type]spacingRule{
		reflect.TypeOf(&Config{}):       {gapWithinType: false},
		reflect.TypeOf(&Secret{}):       {gapWithinType: false, skipGapAfterTypes: []reflect.Type{reflect.TypeOf(&Config{})}},
		reflect.TypeOf(&Database{}):     {gapWithinType: false},
		reflect.TypeOf(&Topic{}):        {gapWithinType: false},
		reflect.TypeOf(&Subscription{}): {gapWithinType: false, skipGapAfterTypes: []reflect.Type{reflect.TypeOf(&Topic{})}},
		reflect.TypeOf(&Enum{}):         {gapWithinType: true},
		reflect.TypeOf(&Data{}):         {gapWithinType: true},
		reflect.TypeOf(&Verb{}):         {gapWithinType: true},
	}

	lastTypePrinted := optional.None[reflect.Type]()
	for _, decl := range m.Decls {
		t := reflect.TypeOf(decl)
		rules, ok := typeSpacingRules[t]
		if !ok {
			rules = spacingRule{gapWithinType: true}
		}
		if lastType, ok := lastTypePrinted.Get(); ok {
			if lastType == t {
				if rules.gapWithinType {
					fmt.Fprintln(w)
				}
			} else if !slices.Contains(rules.skipGapAfterTypes, lastType) {
				fmt.Fprintln(w)
			}
		}
		fmt.Fprintln(w, indent(decl.String()))
		lastTypePrinted = optional.Some[reflect.Type](t)
	}
	fmt.Fprintln(w, "}")
	return w.String()
}

// AddDecls appends decls to the module.
//
// Decls are only added if they are not already present in the module or if they change the visibility of an existing
// Decl.
func (m *Module) AddDecls(decls []Decl) {
	// decls are namespaced by their type.
	typeQualifiedName := func(d Decl) string {
		return reflect.TypeOf(d).Name() + "." + d.GetName()
	}

	existingDecls := map[string]Decl{}
	for _, d := range m.Decls {
		existingDecls[typeQualifiedName(d)] = d
	}
	for _, newDecl := range decls {
		tqName := typeQualifiedName(newDecl)
		if existingDecl, ok := existingDecls[tqName]; ok {
			if newDecl.IsExported() && !existingDecl.IsExported() {
				existingDecls[tqName] = newDecl
			}
			continue
		}

		existingDecls[tqName] = newDecl
	}
	m.Decls = maps.Values(existingDecls)
}

// AddDecl adds a single decl to the module.
//
// It is only added if not already present or if it changes the visibility of the existing Decl.
func (m *Module) AddDecl(decl Decl) {
	m.AddDecls([]Decl{decl})
}

// AddData and return its index.
//
// If data is already in the module, the existing index is returned.
// If the new data is exported but the existing data is not, it sets it to being exported.
func (m *Module) AddData(data *Data) int {
	for i, d := range m.Decls {
		if d, ok := d.(*Data); ok && d.Name == data.Name {
			d.Export = d.Export || data.Export
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
	_ = Visit(m, func(n Node, next func() error) error { //nolint:errcheck
		switch n := n.(type) {
		case *Ref:
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

func (m *Module) GetName() string  { return m.Name }
func (m *Module) IsExported() bool { return false }

// ModuleFromProtoFile loads a module from the given proto-encoded file.
func ModuleFromProtoFile(filename string) (*Module, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ModuleFromBytes(data)
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
