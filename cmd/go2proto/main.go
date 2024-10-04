package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"maps"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"text/template"

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

type Config struct {
	Output    string   `help:"Output file to write generated protobuf schema to." short:"o"`
	Imports   []string `help:"Additional imports to include in the generated protobuf schema." short:"I"`
	GoPackage string   `help:"Go package to use in the generated protobuf schema." short:"g"`

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
	err = generate(out, cli, resolved)
	if gerr := new(GenError); errors.As(err, &gerr) {
		pos := fset.Position(gerr.pos)
		kctx.Fatalf("%s:%d: %s", pos.Filename, pos.Line, err)
	} else {
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
	Config
	Decls   map[string]string
	Renamed map[string]string
	*PkgRefs
}

func (s State) String() string {
	w := &strings.Builder{}
	for _, name := range slices.Sorted(maps.Keys(s.Decls)) {
		w.WriteString(s.Decls[name])
	}
	return w.String()
}

func genErrorf(pos token.Pos, format string, args ...any) error {
	err := fmt.Errorf(format, args...)
	if gerr := new(GenError); errors.As(err, &gerr) {
		return &GenError{pos: gerr.pos, err: err}
	}
	return &GenError{pos: pos, err: err}
}

var tmpl = template.Must(template.New("proto").
	Parse(`
syntax = "proto3";

package {{ .Package }};
{{ range .Imports }}
import "{{.}}";
{{- end}}
{{ if .GoPackage }}
option go_package = "{{ .GoPackage }}";
{{ end -}}
option java_multiple_files = true;

{{ range $name, $decl := .Decls }}
{{- $decl }}
{{ end}}
`))

func generate(out *os.File, config Config, pkg *PkgRefs) error {
	state := State{
		Config:  config,
		Decls:   map[string]string{},
		Renamed: map[string]string{},
		PkgRefs: pkg,
	}
	for _, sym := range pkg.Refs {
		obj := pkg.Pkg.Types.Scope().Lookup(sym)
		if obj == nil {
			return fmt.Errorf("%s: not found in package %s", sym, pkg.Pkg.ID)
		}
		if err := state.generateType(obj, obj.Type()); err != nil {
			return fmt.Errorf("%s: %w", sym, err)
		}
	}
	if err := tmpl.Execute(out, state); err != nil {
		return fmt.Errorf("template error: %w", err)
	}
	return nil
}

func (s *State) resolve(name string) (resolvedName string, ok bool) {
	resolvedName, ok = s.Renamed[name]
	if ok {
		name = resolvedName
	}
	_, ok = s.Decls[name]
	return name, ok
}

func (s *State) addImport(name string) {
	for _, imp := range s.Imports {
		if imp == name {
			return
		}
	}
	s.Imports = append(s.Imports, name)
}

func (s *State) generateType(obj types.Object, t types.Type) error {
	switch t := t.(type) {
	case *types.Named:
		if t.TypeParams() != nil {
			return genErrorf(obj.Pos(), "generic types are not supported")
		}
		switch u := t.Underlying().(type) {
		case *types.Struct:
			if err := s.extractStruct(t, u); err != nil {
				return genErrorf(obj.Pos(), "%w", err)
			}
			return nil

		case *types.Interface:
			return s.extractSumType(t.Obj(), u)

		case *types.Basic:
			return s.extractEnum(t)

		default:
			return genErrorf(obj.Pos(), "unsupported named type %T", u)
		}

	case *types.Basic:
		return nil

	case *types.Slice:
		return s.generateType(obj, t.Elem())

	case *types.Pointer:
		return s.generateType(obj, t.Elem())

	case *types.Interface:
		return genErrorf(obj.Pos(), "unnamed interfaces are not supported")

	default:
		return genErrorf(obj.Pos(), "unsupported type %T", obj.Type())
	}
}

type builtinType struct {
	ref  string
	path string
}

var builtinTypes = map[string]builtinType{
	"time.Time":     {"google.protobuf.Timestamp", "google/protobuf/timestamp.proto"},
	"time.Duration": {"google.protobuf.Duration", "google/protobuf/duration.proto"},
}

func (s *State) extractStruct(n *types.Named, t *types.Struct) error {
	if imp, ok := builtinTypes[n.String()]; ok {
		s.addImport(imp.path)
		return nil
	}

	name, ok := s.resolve(n.Obj().Name())
	if ok {
		return nil
	}
	s.Decls[name] = ""
	w := &strings.Builder{}
	fmt.Fprintf(w, "message %s {\n", name)
	for i := range t.NumFields() {
		field := t.Field(i)
		pb := reflect.StructTag(t.Tag(i)).Get("protobuf")
		if pb == "-" {
			continue
		} else if pb == "" {
			return genErrorf(n.Obj().Pos(), "%s: missing protobuf tag", field.Name())
		}
		tag, err := parsePBTag(pb)
		if err != nil {
			return genErrorf(n.Obj().Pos(), "%s: %w", field.Name(), err)
		}
		prefix := ""
		if tag.Optional {
			prefix = "optional "
		}
		if err := s.generateType(field, field.Type()); err != nil {
			return fmt.Errorf("%s: %w", field.Name(), err)
		}
		fmt.Fprintf(w, "  %s%s %s = %d;\n", prefix, typeRef(field.Type()), strcase.ToLowerSnake(field.Name()), tag.ID)
	}
	fmt.Fprintf(w, "}\n")
	s.Decls[name] = w.String()
	return nil
}

func (s *State) extractSumType(obj types.Object, i *types.Interface) error {
	sumTypeName, ok := s.resolve(obj.Name())
	if ok {
		return nil
	}
	s.Decls[sumTypeName] = ""
	w := &strings.Builder{}
	sums := map[string]int{}
	fmt.Fprintf(w, "message %s {\n", sumTypeName)
	fmt.Fprintf(w, "  oneof value {\n")
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
			if err := s.generateType(sym, sym.Type()); err != nil {
				return genErrorf(sym.Pos(), "%s: %w", name, err)
			}
			sums[name] = directive.ID
		}
	}
	// The ID's we generate here aren't stable. Not sure what to do about that, but for now we just sort them and deal with
	// the backwards incompatibility. The buf linter will pick this up in PRs though.
	for _, sum := range slices.Sorted(maps.Keys(sums)) {
		fieldName := strcase.ToLowerCamel(strings.TrimPrefix(sum, sumTypeName))
		fmt.Fprintf(w, "    %s %s = %d;\n", sum, fieldName, sums[sum])
	}
	fmt.Fprintf(w, "  }\n")
	fmt.Fprintf(w, "}\n")
	s.Decls[sumTypeName] = w.String()
	return nil
}

func (s *State) extractEnum(t *types.Named) error {
	enumName, ok := s.resolve(t.Obj().Name())
	if ok {
		return nil
	}
	s.Decls[enumName] = ""
	w := &strings.Builder{}
	enums := map[string]int{}
	fmt.Fprintf(w, "enum %s {\n", enumName)
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
		enums[name] = n
	}
	for i, sum := range slices.Sorted(maps.Keys(enums)) {
		fmt.Fprintf(w, "  %s = %d;\n", strcase.ToUpperSnake(sum), i)
	}
	fmt.Fprintf(w, "}\n")
	s.Decls[enumName] = w.String()
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

func typeRef(t types.Type) string {
	switch t := t.(type) {
	case *types.Named:
		ref := t.Obj().Pkg().Path() + "." + t.Obj().Name()
		if t, ok := builtinTypes[ref]; ok {
			return t.ref
		}
		return t.Obj().Name()

	case *types.Slice:
		if t.Elem().String() == "byte" {
			return "bytes"
		}
		return "repeated " + typeRef(t.Elem())

	case *types.Pointer:
		return typeRef(t.Elem())

	default:
		switch t.String() {
		case "int":
			return "int64"

		case "uint":
			return "uint64"

		case "float64":
			return "double"

		case "float32":
			return "float"

		case "string", "bool", "uint64", "int64", "uint32", "int32":
			return t.String()

		default:
			panic(fmt.Sprintf("unsupported type %s", t.String()))

		}
	}
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
