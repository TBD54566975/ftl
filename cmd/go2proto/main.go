package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
	"golang.org/x/tools/go/packages"

	"github.com/TBD54566975/ftl/internal/schema/strcase"
)

const help = `Generate a Protobuf schema from Go types.

It supports converting structs to messages, Go "sum types" to oneof fields, and Go "enums" to Protobuf enums.

The generator works by extracting protobuf tags from the source Go types. There are two locations where these tags must
be specified:

  1. For fields using a tag in the form ` + "`protobuf:\"<id>[,optional]\"`" + `.
  2. For sum types as comment directives in the form //protobuf:<id>.

An example showing all three supported types and the corresponding protobuf tags:

	type UserType int

	const (
		UserTypeUnknown UserType = iota
		UserTypeAdmin
		UserTypeUser
	)

	// Entity is a "sum type" consisting of User and Group.
	//
	// Every sum type element must have a comment directive in the form //protobuf:<id>.
	type Entity interface { entity() }

	//protobuf:1
	type User struct {
		Name   string    ` + "`protobuf:\"1\"`" + `
		Type   UserType ` + "`protobuf:\"2\"`" + `
	}
	func (User) entity() {}

	//protobuf:2
	type Group struct {
		Users []string ` + "`protobuf:\"1\"`" + `
	}
	func (Group) entity() {}

	type Role struct {
		Name string ` + "`protobuf:\"1\"`" + `
		Entities []Entity ` + "`protobuf:\"2\"`" + `
	}

And this is the corresponding protobuf schema:

	message Entity {
	  oneof value {
	    User user = 1;
	    Group group = 2;
	  }
	}

	enum UserType {
	  USER_TYPE_UNKNOWN = 0;
	  USER_TYPE_ADMIN = 1;
	  USER_TYPE_USER = 2;
	}

	message User {
	  string Name = 1;
	  UserType Type = 2;
	}

	message Group {
	  repeated string users = 1;
	}

	message Role {
	  string name = 1;
	  repeated Entity entities = 2;
	}
`

type File struct {
	GoPackage string
	Imports   []string
	Decls     []Decl
}

func (f *File) AddImport(name string) {
	for _, imp := range f.Imports {
		if imp == name {
			return
		}
	}
	f.Imports = append(f.Imports, name)
}

func (f File) OrderedDecls() []Decl {
	decls := make([]Decl, len(f.Decls))
	copy(decls, f.Decls)
	sort.Slice(decls, func(i, j int) bool {
		return decls[i].DeclName() < decls[j].DeclName()
	})
	return decls
}

func (f File) TypeOf(name string) string {
	for _, decl := range f.Decls {
		if decl.DeclName() == name {
			return reflect.Indirect(reflect.ValueOf(decl)).Type().Name()
		}
	}
	panic("unknown type " + name)
}

//sumtype:decl
type Decl interface {
	decl()
	DeclName() string
}

type Message struct {
	Name   string
	Fields []Field
}

func (Message) decl()              {}
func (m Message) DeclName() string { return m.Name }

type Field struct {
	ID          int
	Name        string
	Type        string
	Optional    bool
	Repeated    bool
	ProtoGoType string
	Pointer     bool
}

var reservedWords = map[string]string{
	"String": "String_",
}

func protoName(s string) string {
	if name, ok := reservedWords[s]; ok {
		return name
	}
	return strcase.ToUpperCamel(s)
}

func (f Field) EscapedName() string {
	if name, ok := reservedWords[f.Name]; ok {
		return name
	}
	return strcase.ToUpperCamel(f.Name)
}

type Enum struct {
	Name   string
	Values map[string]int
}

func (e Enum) ByValue() map[int]string {
	m := map[int]string{}
	for k, v := range e.Values {
		m[v] = k
	}
	return m
}

func (Enum) decl()              {}
func (e Enum) DeclName() string { return e.Name }

type SumType struct {
	Name     string
	Variants map[string]int
}

func (SumType) decl()              {}
func (s SumType) DeclName() string { return s.Name }

type Config struct {
	Output  string            `help:"Output file to write generated protobuf schema to." short:"o" xor:"output"`
	JSON    bool              `help:"Dump intermediate JSON represesentation." short:"j" xor:"output"`
	Options map[string]string `placeholder:"OPTION=VALUE" help:"Additional options to include in the generated protobuf schema. Note: strings must be double quoted." short:"O" mapsep:"\\0"`
	Mappers bool              `help:"Generate ToProto and FromProto mappers for each message." short:"m"`

	Package string   `arg:"" help:"Package name to use in the generated protobuf schema."`
	Ref     []string `arg:"" help:"Type to generate protobuf schema from in the form PKG.TYPE. eg. github.com/foo/bar/waz.Waz or ./waz.Waz" required:"true" placeholder:"PKG.TYPE"`
}

func main() {
	fset := token.NewFileSet()
	cli := Config{}
	kctx := kong.Parse(&cli, kong.Description(help), kong.UsageOnError())

	out := os.Stdout
	if cli.Output != "" {
		var err error
		out, err = os.Create(cli.Output + "~")
		kctx.FatalIfErrorf(err)
		defer out.Close()
		defer os.Remove(cli.Output + "~")
	}

	var resolved *PkgRefs
	for _, ref := range cli.Ref {
		parts := strings.Split(ref, ".")
		pkg := strings.Join(parts[:len(parts)-1], ".")
		if resolved != nil && resolved.Path != pkg {
			kctx.Fatalf("only a single package is supported")
		} else if resolved == nil {
			resolved = &PkgRefs{Ref: ref, Path: pkg}
		}
		resolved.Refs = append(resolved.Refs, parts[len(parts)-1])
	}
	pkgs, err := packages.Load(&packages.Config{
		Fset: fset,
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedImports | packages.NeedSyntax |
			packages.NeedFiles | packages.NeedName,
	}, resolved.Path)
	kctx.FatalIfErrorf(err)
	commentMap := ast.CommentMap{}
	for _, pkg := range pkgs {
		resolved.Pkg = pkg
		if len(pkg.Errors) > 0 {
			fmt.Fprintf(os.Stderr, "go2proto: warning: %s\n", pkg.Errors[0])
			break
		}
		for _, file := range pkg.Syntax {
			fcmap := ast.NewCommentMap(fset, file, file.Comments)
			maps.Copy(commentMap, fcmap)
		}
	}
	resolved.Comments = commentMap
	if resolved.Pkg.Types == nil {
		kctx.Fatalf("package %s had fatal errors, cannot continue", resolved.Path)
	}
	file, err := extract(cli, resolved)
	if gerr := new(GenError); errors.As(err, &gerr) {
		pos := fset.Position(gerr.pos)
		kctx.Fatalf("%s:%d: %s", pos.Filename, pos.Line, err)
	} else {
		kctx.FatalIfErrorf(err)
	}

	if cli.JSON {
		b, err := json.MarshalIndent(file, "", "  ")
		kctx.FatalIfErrorf(err)
		fmt.Println(string(b))
		return
	}

	err = render(out, cli, file)
	kctx.FatalIfErrorf(err)

	if cli.Mappers {
		w, err := os.CreateTemp(resolved.Path, "go2proto.to.go-*")
		kctx.FatalIfErrorf(err)
		defer os.Remove(w.Name())
		defer w.Close()
		err = renderToProto(w, cli, file)
		kctx.FatalIfErrorf(err)
		err = os.Rename(w.Name(), filepath.Join(resolved.Path, "go2proto.to.go"))
		kctx.FatalIfErrorf(err)
	}

	if cli.Output != "" {
		err = os.Rename(cli.Output+"~", cli.Output)
	}
	kctx.FatalIfErrorf(err)
}

type GenError struct {
	pos token.Pos
	err error
}

func (g GenError) Error() string { return g.err.Error() }
func (g GenError) Unwrap() error { return g.err }

type PkgRefs struct {
	Comments ast.CommentMap
	Path     string
	Ref      string
	Refs     []string
	Pkg      *packages.Package
}

type State struct {
	Dest File
	Seen map[string]bool
	Config
	*PkgRefs
}

func genErrorf(pos token.Pos, format string, args ...any) error {
	err := fmt.Errorf(format, args...)
	if gerr := new(GenError); errors.As(err, &gerr) {
		return &GenError{pos: gerr.pos, err: err}
	}
	return &GenError{pos: pos, err: err}
}

func extract(config Config, pkg *PkgRefs) (File, error) {
	state := State{
		Seen:    map[string]bool{},
		Config:  config,
		PkgRefs: pkg,
	}
	for _, sym := range pkg.Refs {
		obj := pkg.Pkg.Types.Scope().Lookup(sym)
		if obj == nil {
			return File{}, fmt.Errorf("%s: not found in package %s", sym, pkg.Pkg.ID)
		}
		if !strings.HasSuffix(pkg.Pkg.Name, "_test") {
			state.Dest.GoPackage = pkg.Pkg.Name
		}
		if err := state.extractDecl(obj, obj.Type()); err != nil {
			return File{}, fmt.Errorf("%s: %w", sym, err)
		}
	}
	return state.Dest, nil
}

func (s *State) extractDecl(obj types.Object, t types.Type) error {
	named, ok := t.(*types.Named)
	if !ok {
		return genErrorf(obj.Pos(), "expected named type, got %T", t)
	}
	if named.TypeParams() != nil {
		return genErrorf(obj.Pos(), "generic types are not supported")
	}
	switch u := t.Underlying().(type) {
	case *types.Struct:
		if err := s.extractStruct(named, u); err != nil {
			return genErrorf(obj.Pos(), "%w", err)
		}
		return nil

	case *types.Interface:
		return s.extractSumType(named.Obj(), u)

	case *types.Basic:
		return s.extractEnum(named)

	default:
		return genErrorf(obj.Pos(), "unsupported named type %T", u)
	}
}

type stdType struct {
	ref  string
	path string
}

var stdTypes = map[string]stdType{
	"time.Time":     {"google.protobuf.Timestamp", "google/protobuf/timestamp.proto"},
	"time.Duration": {"google.protobuf.Duration", "google/protobuf/duration.proto"},
}

func (s *State) extractStruct(n *types.Named, t *types.Struct) error {
	if imp, ok := stdTypes[n.String()]; ok {
		s.Dest.AddImport(imp.path)
		return nil
	}

	name := n.Obj().Name()
	if _, ok := s.Seen[name]; ok {
		return nil
	}
	s.Seen[name] = true
	decl := &Message{
		Name: name,
	}
	for i := range t.NumFields() {
		rf := t.Field(i)
		pb := reflect.StructTag(t.Tag(i)).Get("protobuf")
		if pb == "-" {
			continue
		} else if pb == "" {
			return genErrorf(n.Obj().Pos(), "%s: missing protobuf tag", rf.Name())
		}
		field := Field{
			Name: rf.Name(),
		}
		if err := s.applyFieldType(rf.Type(), &field); err != nil {
			return fmt.Errorf("%s: %w", rf.Name(), err)
		}
		tag, err := parsePBTag(pb)
		if err != nil {
			return genErrorf(n.Obj().Pos(), "%s: %w", rf.Name(), err)
		}
		field.ID = tag.ID
		field.Optional = tag.Optional
		decl.Fields = append(decl.Fields, field)
		if field.Optional && field.Repeated {
			return genErrorf(n.Obj().Pos(), "%s: repeated optional fields are not supported", rf.Name())
		}
		if nt, ok := rf.Type().(*types.Named); ok {
			if err := s.extractDecl(rf, nt); err != nil {
				return fmt.Errorf("%s: %w", rf.Name(), err)
			}
		}
	}
	s.Dest.Decls = append(s.Dest.Decls, decl)
	return nil
}

func (s *State) extractSumType(obj types.Object, i *types.Interface) error {
	sumTypeName := obj.Name()
	if _, ok := s.Seen[sumTypeName]; ok {
		return nil
	}
	s.Seen[sumTypeName] = true
	decl := SumType{
		Name:     sumTypeName,
		Variants: map[string]int{},
	}
	scope := s.Pkg.Types.Scope()
	for _, name := range scope.Names() {
		sym := scope.Lookup(name)
		if sym == obj {
			continue
		}
		if types.Implements(sym.Type(), i) || types.Implements(types.NewPointer(sym.Type()), i) {
			var directive *pbTag
			if comments := findCommentsForObject(sym, s.Pkg.Syntax); comments != nil {
				for _, line := range comments.List {
					if strings.HasPrefix(line.Text, "//protobuf:") {
						tag, err := parsePBTag(strings.TrimPrefix(line.Text, "//protobuf:"))
						if err != nil {
							return genErrorf(sym.Pos(), "invalid //protobuf: directive %q: %w", line.Text, err)
						}
						directive = &tag
					}
				}
			}
			if directive == nil {
				return genErrorf(sym.Pos(), "sum type element is missing //protobuf:<id> directive")
			}
			if err := s.extractDecl(sym, sym.Type()); err != nil {
				return genErrorf(sym.Pos(), "%s: %w", name, err)
			}
			decl.Variants[name] = directive.ID
		}
	}
	s.Dest.Decls = append(s.Dest.Decls, decl)
	return nil
}

func (s *State) extractEnum(t *types.Named) error {
	if imp, ok := stdTypes[t.String()]; ok {
		s.Dest.AddImport(imp.path)
		return nil
	}
	enumName := t.Obj().Name()
	if _, ok := s.Seen[enumName]; ok {
		return nil
	}
	s.Seen[enumName] = true
	decl := Enum{
		Name:   enumName,
		Values: map[string]int{},
	}
	scope := s.Pkg.Types.Scope()
	for _, name := range scope.Names() {
		sym := scope.Lookup(name)
		if sym == t.Obj() || sym.Type() != t {
			continue
		}
		c, ok := sym.(*types.Const)
		if !ok {
			return genErrorf(sym.Pos(), "expected const")
		}
		n, err := strconv.Atoi(c.Val().String())
		if err != nil {
			return genErrorf(sym.Pos(), "enum value %q must be a constant integer: %w", c.Val(), err)
		}
		if strcase.ToUpperCamel(name) != name {
			return genErrorf(sym.Pos(), "enum value %q must be upper camel case %q", name, strcase.ToUpperCamel(name))
		}
		if !strings.HasPrefix(name, enumName) {
			return genErrorf(sym.Pos(), "enum value %q must start with %q", name, enumName)
		}
		decl.Values[name] = n
	}
	s.Dest.Decls = append(s.Dest.Decls, decl)
	return nil
}

type pbTag struct {
	ID       int
	Optional bool
}

func parsePBTag(tag string) (pbTag, error) {
	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		return pbTag{}, fmt.Errorf("missing tag")
	}
	id, err := strconv.Atoi(parts[0])
	if err != nil {
		return pbTag{}, fmt.Errorf("invalid id: %w", err)
	}
	out := pbTag{ID: id}
	for _, part := range parts[1:] {
		switch part {
		case "optional":
			out.Optional = true

		default:
			return pbTag{}, fmt.Errorf("unknown tag: %s", tag)
		}
	}
	return out, nil
}

func (s *State) applyFieldType(t types.Type, field *Field) error {
	switch t := t.(type) {
	case *types.Named:
		if err := s.extractDecl(t.Obj(), t); err != nil {
			return err
		}
		ref := t.Obj().Pkg().Path() + "." + t.Obj().Name()
		if bt, ok := stdTypes[ref]; ok {
			field.Type = bt.ref
		} else {
			field.Type = t.Obj().Name()
		}

	case *types.Slice:
		if t.Elem().String() == "byte" {
			field.Type = "bytes"
		} else {
			field.Repeated = true
			return s.applyFieldType(t.Elem(), field)
		}

	case *types.Pointer:
		field.Pointer = true
		if _, ok := t.Elem().(*types.Slice); ok {
			return fmt.Errorf("pointer to named type is not supported")
		}
		return s.applyFieldType(t.Elem(), field)

	default:
		field.ProtoGoType = t.String()
		switch t.String() {
		case "int":
			field.Type = "int64"
			field.ProtoGoType = "int64"

		case "uint":
			field.Type = "uint64"
			field.ProtoGoType = "uint64"

		case "float64":
			field.Type = "double"
			field.ProtoGoType = "float64"

		case "float32":
			field.Type = "float"
			field.ProtoGoType = "float32"

		case "string", "bool", "uint64", "int64", "uint32", "int32":
			field.Type = t.String()

		default:
			return fmt.Errorf("unsupported type %s", t.String())

		}
	}
	return nil
}

func findCommentsForObject(obj types.Object, syntax []*ast.File) *ast.CommentGroup {
	for _, file := range syntax {
		if file.Pos() <= obj.Pos() && obj.Pos() <= file.End() {
			// Use ast.Inspect to traverse the AST and locate the node
			var comments *ast.CommentGroup
			ast.Inspect(file, func(n ast.Node) bool {
				if n == nil {
					return false
				}
				// If found, get the documentation comments
				if node, ok := n.(*ast.GenDecl); ok {
					for _, spec := range node.Specs {
						if spec.Pos() == obj.Pos() {
							comments = node.Doc
							return false // Stop the traversal once the node is found
						}
					}
				}
				return true
			})
			return comments
		}
	}
	return nil
}
