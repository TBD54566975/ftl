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
	"time"

	"github.com/alecthomas/types/optional"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
)

type ModuleRuntime struct {
	CreateTime  time.Time `protobuf:"1"`
	Language    string    `protobuf:"2"`
	MinReplicas int32     `protobuf:"3"`
	OS          string    `protobuf:"4,optional"`
	Arch        string    `protobuf:"5,optional"`
	// Image is the name of the runner image. Defaults to "ftl0/ftl-runner".
	// Must not include a tag, as FTL's version will be used as the tag.
	Image string `protobuf:"6,optional"`
}

var _ Runtime = (*ModuleRuntime)(nil)

func (m *ModuleRuntime) runtime() {}

type Module struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" protobuf:"2"`
	Builtin  bool     `parser:"@'builtin'?" protobuf:"3"`
	Name     string   `parser:"'module' @Ident '{'" protobuf:"4"`
	Decls    []Decl   `parser:"@@* '}'" protobuf:"5"`

	Runtime *ModuleRuntime `protobuf:"31634,optional" parser:""`
}

var _ Node = (*Module)(nil)
var _ Symbol = (*Module)(nil)
var _ sql.Scanner = (*Module)(nil)
var _ driver.Valuer = (*Module)(nil)

func (m *Module) Value() (driver.Value, error) {
	value, err := proto.Marshal(Redact(m).ToProto())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Module to proto: %w", err)
	}
	return value, nil
}
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
		reflect.TypeOf(&Config{}):    {gapWithinType: false},
		reflect.TypeOf(&Secret{}):    {gapWithinType: false, skipGapAfterTypes: []reflect.Type{reflect.TypeOf(&Config{})}},
		reflect.TypeOf(&Database{}):  {gapWithinType: false},
		reflect.TypeOf(&Topic{}):     {gapWithinType: false},
		reflect.TypeOf(&TypeAlias{}): {gapWithinType: false},
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
	var runtime *schemapb.ModuleRuntime
	if m.Runtime != nil {
		runtime = &schemapb.ModuleRuntime{
			CreateTime:  timestamppb.New(m.Runtime.CreateTime),
			Language:    m.Runtime.Language,
			MinReplicas: m.Runtime.MinReplicas,
		}
		if m.Runtime.OS != "" {
			runtime.Os = &m.Runtime.OS
		}
		if m.Runtime.Arch != "" {
			runtime.Arch = &m.Runtime.Arch
		}
		if m.Runtime.Image != "" {
			runtime.Image = &m.Runtime.Image
		}
	}
	return &schemapb.Module{
		Pos:      posToProto(m.Pos),
		Builtin:  m.Builtin,
		Name:     m.Name,
		Comments: m.Comments,
		Decls:    declListToProto(m.Decls),
		Runtime:  runtime,
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
		Pos:      PosFromProto(s.Pos),
		Builtin:  s.Builtin,
		Name:     s.Name,
		Comments: s.Comments,
		Decls:    declListToSchema(s.Decls),
	}
	if s.Runtime != nil {
		module.Runtime = &ModuleRuntime{
			CreateTime:  s.Runtime.GetCreateTime().AsTime(),
			Language:    s.Runtime.Language,
			MinReplicas: s.Runtime.MinReplicas,
			OS:          s.Runtime.GetOs(),
			Arch:        s.Runtime.GetArch(),
			Image:       s.Runtime.GetImage(),
		}
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
