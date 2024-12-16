package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"iter"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
	"golang.org/x/tools/go/packages"

	"github.com/block/ftl/common/strcase"
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

var (
	textMarshaler   = loadInterface("encoding", "TextMarshaler")
	binaryMarshaler = loadInterface("encoding", "BinaryMarshaler")
	// stdTypes is a map of Go types to corresponding protobuf types.
	stdTypes = map[string]stdType{
		"time.Time":     {"google.protobuf.Timestamp", "google/protobuf/timestamp.proto"},
		"time.Duration": {"google.protobuf.Duration", "google/protobuf/duration.proto"},
	}
	builtinTypes = map[string]struct{}{
		"bool": {}, "int": {}, "int8": {}, "int16": {}, "int32": {}, "int64": {}, "uint": {}, "uint8": {}, "uint16": {},
		"uint32": {}, "uint64": {}, "float32": {}, "float64": {}, "string": {},
	}
)

type stdType struct {
	ref  string
	path string
}

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

// KindOf looks up the kind of a type in the declarations. Returns KindUnspecified if the type is not found.
func (f File) KindOf(name string) Kind {
	for _, decl := range f.Decls {
		if decl.DeclName() == name {
			return Kind(reflect.Indirect(reflect.ValueOf(decl)).Type().Name())
		}
	}
	if _, ok := builtinTypes[name]; ok {
		return KindBuiltin
	}
	if _, ok := stdTypes[name]; ok {
		return KindStdlib
	}
	return KindUnspecified
}

//sumtype:decl
type Decl interface {
	decl()
	DeclName() string
}

type Message struct {
	Comment string
	Name    string
	Fields  []*Field
}

func (Message) decl()              {}
func (m Message) DeclName() string { return m.Name }

type Kind string

const (
	KindUnspecified     Kind = ""
	KindBuiltin         Kind = "Builtin"
	KindStdlib          Kind = "Stdlib"
	KindMessage         Kind = "Message"
	KindEnum            Kind = "Enum"
	KindSumType         Kind = "SumType"
	KindBinaryMarshaler Kind = "BinaryMarshaler"
	KindTextMarshaler   Kind = "TextMarshaler"
)

type Field struct {
	ID          int
	Name        string
	OriginType  string // The original type of the field, eg. int, string, float32, etc.
	ProtoType   string // The type of the field in the generated .proto file.
	ProtoGoType string // The type of the field in the generated Go protobuf code. eg. int -> int64.
	Optional    bool
	Repeated    bool
	Pointer     bool

	Kind Kind
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
	Comment string
	Name    string
	Values  map[string]int
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
	Comment  string
	Name     string
	Variants map[string]int
}

func (SumType) decl()              {}
func (s SumType) DeclName() string { return s.Name }

// TextMarshaler is a named type that implements encoding.TextMarshaler. Encoding will delegate to the marshaller.
type TextMarshaler struct {
	Comment string
	Name    string
}

func (TextMarshaler) decl()              {}
func (u TextMarshaler) DeclName() string { return u.Name }

// BinaryMarshaler is a named type that implements encoding.BinaryMarshaler. Encoding will delegate to the marshaller.
type BinaryMarshaler struct {
	Comment string
	Name    string
}

func (BinaryMarshaler) decl()              {}
func (u BinaryMarshaler) DeclName() string { return u.Name }

type Config struct {
	Output  string            `help:"Output file to write generated protobuf schema to." short:"o" xor:"output"`
	JSON    bool              `help:"Dump intermediate JSON represesentation." short:"j" xor:"output"`
	Options map[string]string `placeholder:"OPTION=VALUE" help:"Additional options to include in the generated protobuf schema. Note: strings must be double quoted." short:"O" mapsep:"\\0"`
	Mappers bool              `help:"Generate ToProto and FromProto mappers for each message." short:"m"`

	Package string   `arg:"" help:"Package name to use in the generated protobuf schema."`
	Ref     []string `arg:"" help:"Type to generate protobuf schema from in the form PKG.TYPE. eg. github.com/foo/bar/waz.Waz or ./waz.Waz" required:"true" placeholder:"PKG.TYPE"`
}

func main() {
	cli := Config{}
	kctx := kong.Parse(&cli, kong.Description(help), kong.UsageOnError())
	err := run(cli)
	kctx.FatalIfErrorf(err)
}

func run(cli Config) error {
	out := os.Stdout
	if cli.Output != "" {
		var err error
		out, err = os.Create(cli.Output + "~")
		if err != nil {
			return fmt.Errorf("")
		}
		defer out.Close()
		defer os.Remove(cli.Output + "~")
	}

	var resolved *PkgRefs
	for _, ref := range cli.Ref {
		parts := strings.Split(ref, ".")
		pkg := strings.Join(parts[:len(parts)-1], ".")
		if resolved != nil && resolved.Path != pkg {
			return fmt.Errorf("only a single package is supported")
		} else if resolved == nil {
			resolved = &PkgRefs{Ref: ref, Path: pkg}
		}
		resolved.Refs = append(resolved.Refs, parts[len(parts)-1])
	}
	fset := token.NewFileSet()
	pkgs, err := packages.Load(&packages.Config{
		Fset: fset,
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedImports | packages.NeedSyntax |
			packages.NeedFiles | packages.NeedName,
	}, resolved.Path)
	if err != nil {
		return fmt.Errorf("unable to load package %s: %w", resolved.Path, err)
	}
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
		return fmt.Errorf("package %s had fatal errors, cannot continue", resolved.Path)
	}
	file, err := extract(cli, resolved)
	if gerr := new(GenError); errors.As(err, &gerr) {
		pos := fset.Position(gerr.pos)
		return fmt.Errorf("%s:%d: %w", pos.Filename, pos.Line, err)
	} else {
		if err != nil {
			return err
		}
	}

	if cli.JSON {
		b, err := json.MarshalIndent(file, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(b))
		return nil
	}

	err = render(out, cli, file)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	if cli.Mappers {
		w, err := os.CreateTemp(resolved.Path, "go2proto.to.go-*")
		if err != nil {
			return fmt.Errorf("create temp: %w", err)
		}
		defer os.Remove(w.Name())
		defer w.Close()
		err = renderToProto(w, cli, file)
		if err != nil {
			return err
		}
		err = os.Rename(w.Name(), filepath.Join(resolved.Path, "go2proto.to.go"))
		if err != nil {
			return fmt.Errorf("rename: %w", err)
		}
	}

	if cli.Output != "" {
		err = os.Rename(cli.Output+"~", cli.Output)
		if err != nil {
			return fmt.Errorf("rename: %w", err)
		}
	}
	return nil
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
	Pass     int
	Messages map[*Message]*types.Named
	Dest     File
	Seen     map[string]bool
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
		Messages: map[*Message]*types.Named{},
		Seen:     map[string]bool{},
		Config:   config,
		PkgRefs:  pkg,
	}
	// First pass, extract all the decls.
	for _, sym := range pkg.Refs {
		obj := pkg.Pkg.Types.Scope().Lookup(sym)
		if obj == nil {
			return File{}, fmt.Errorf("%s: not found in package %s", sym, pkg.Pkg.ID)
		}
		if !strings.HasSuffix(pkg.Pkg.Name, "_test") {
			state.Dest.GoPackage = pkg.Pkg.Name
		}
		named, ok := obj.Type().(*types.Named)
		if !ok {
			return File{}, genErrorf(obj.Pos(), "%s: expected named type, got %T", sym, obj.Type())
		}
		if err := state.extractDecl(obj, named); err != nil {
			return File{}, fmt.Errorf("%s: %w", sym, err)
		}
	}
	state.Pass++
	// Second pass, populate the fields of messages.
	for msg, n := range state.Messages {
		if err := state.populateFields(msg, n); err != nil {
			return File{}, fmt.Errorf("%s: %w", msg.Name, err)
		}
	}
	return state.Dest, nil
}

func (s *State) extractDecl(obj types.Object, named *types.Named) error {
	if named.TypeParams() != nil {
		return genErrorf(obj.Pos(), "generic types are not supported")
	}
	if imp, ok := stdTypes[named.String()]; ok {
		s.Dest.AddImport(imp.path)
		return nil
	}
	switch u := named.Underlying().(type) {
	case *types.Struct:
		if err := s.extractStruct(named); err != nil {
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

func (s *State) extractStruct(n *types.Named) error {
	name := n.Obj().Name()
	if _, ok := s.Seen[name]; ok {
		return nil
	}
	s.Seen[name] = true
	if implements(n, binaryMarshaler) || implements(n, textMarshaler) {
		return nil
	}
	decl := &Message{
		Name: name,
	}
	if comment := findCommentsForObject(n.Obj(), s.Pkg.Syntax); comment != nil {
		decl.Comment = comment.Text()
	}
	// First pass over structs we just want to extract type information. The fields themselves will be populated in the
	// second pass.
	fields, errf := iterFields(n)
	for rf := range fields {
		if err := s.maybeExtractDecl(rf, rf.Type()); err != nil {
			return err
		}
	}
	if err := errf(); err != nil {
		return err
	}
	s.Messages[decl] = n
	s.Dest.Decls = append(s.Dest.Decls, decl)
	return nil
}

func (s *State) maybeExtractDecl(n types.Object, t types.Type) error {
	switch t := t.(type) {
	case *types.Named:
		return s.extractDecl(n, t)

	case *types.Slice:
		return s.maybeExtractDecl(n, t.Elem())

	case *types.Pointer:
		return s.maybeExtractDecl(n, t.Elem())

	case *types.Interface:
		return s.extractSumType(n, t)

	default:
		return nil
	}
}

func (s *State) populateFields(decl *Message, n *types.Named) error {
	fields, errf := iterFields(n)
	for rf, tag := range fields {
		field := &Field{
			Name: rf.Name(),
		}
		if err := s.applyFieldType(rf.Type(), field); err != nil {
			return fmt.Errorf("%s: %w", rf.Name(), err)
		}
		field.ID = tag.ID
		field.Optional = tag.Optional
		if field.Optional && field.Repeated {
			return genErrorf(n.Obj().Pos(), "%s: repeated optional fields are not supported", rf.Name())
		}
		if nt, ok := rf.Type().(*types.Named); ok {
			if err := s.extractDecl(rf, nt); err != nil {
				return fmt.Errorf("%s: %w", rf.Name(), err)
			}
		}
		if field.Kind == KindUnspecified {
			field.Kind = s.Dest.KindOf(field.OriginType)
		}
		decl.Fields = append(decl.Fields, field)
	}
	return errf()
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
	if comment := findCommentsForObject(obj, s.Pkg.Syntax); comment != nil {
		decl.Comment = comment.Text()
	}
	scope := s.Pkg.Types.Scope()
	for _, name := range scope.Names() {
		sym := scope.Lookup(name)
		if sym == obj || (!types.Implements(sym.Type(), i) && !types.Implements(types.NewPointer(sym.Type()), i)) {
			continue
		}

		var pbDirectives []*pbTag
		interfaceType := false
		if _, ok := sym.Type().Underlying().(*types.Interface); ok {
			interfaceType = true
		}

		if comments := findCommentsForObject(sym, s.Pkg.Syntax); comments != nil {
			for _, line := range comments.List {
				if strings.HasPrefix(line.Text, "//protobuf:") {
					tag, err := parsePBTag(strings.TrimPrefix(line.Text, "//protobuf:"))
					if err != nil {
						return genErrorf(sym.Pos(), "invalid //protobuf: directive %q: %w", line.Text, err)
					}
					pbDirectives = append(pbDirectives, &tag)
				}
			}
		}
		if len(pbDirectives) == 0 {
			// skip interface types. These would result into nested oneofs. We only include leafs as a flat list.
			if !interfaceType {
				return genErrorf(sym.Pos(), "sum type element is missing //protobuf:<id> directive: %s", sym.Name())
			}
		}
		if err := s.extractDecl(sym, sym.Type().(*types.Named)); err != nil { //nolint:forcetypeassert
			return genErrorf(sym.Pos(), "%s: %w", name, err)
		}
		id := -1
		for _, directive := range pbDirectives {
			if id < 0 && directive.SumType == "" {
				id = directive.ID
			} else if directive.SumType == sumTypeName {
				id = directive.ID
			}
		}
		if !interfaceType {
			// we do not want to repeat both sumtypes and their elements
			decl.Variants[name] = id
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
	if comment := findCommentsForObject(t.Obj(), s.Pkg.Syntax); comment != nil {
		decl.Comment = comment.Text()
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
		if !strings.HasPrefix(name, enumName) {
			return genErrorf(sym.Pos(), "enum value %q must start with %q", name, enumName)
		}
		decl.Values[name] = n
	}
	s.Dest.Decls = append(s.Dest.Decls, decl)
	return nil
}

func (s *State) canMarshal(t types.Type, field *Field, name string) bool {
	if _, ok := stdTypes[name]; !ok && s.Dest.KindOf(name) == KindUnspecified {
		if implements(t, textMarshaler) {
			field.ProtoType = "string"
			field.ProtoGoType = "string"
			field.Kind = KindTextMarshaler
			return true
		} else if implements(t, binaryMarshaler) {
			field.ProtoType = "bytes"
			field.ProtoGoType = "bytes"
			field.Kind = KindBinaryMarshaler
			return true
		}
	}
	return false
}

func (s *State) applyFieldType(t types.Type, field *Field) error {
	field.OriginType = t.String()
	switch t := t.(type) {
	case *types.Alias:
		if s.canMarshal(t, field, t.Obj().Name()) {
			return nil
		}

	case *types.Named:
		if s.canMarshal(t, field, t.Obj().Name()) {
			return nil
		}
		if err := s.extractDecl(t.Obj(), t); err != nil {
			return err
		}
		ref := t.Obj().Pkg().Path() + "." + t.Obj().Name()
		if bt, ok := stdTypes[ref]; ok {
			field.ProtoType = bt.ref
			field.ProtoGoType = protoName(bt.ref)
			field.OriginType = t.Obj().Name()
			field.Kind = KindStdlib
		} else {
			field.ProtoType = t.Obj().Name()
			field.ProtoGoType = protoName(t.Obj().Name())
			field.OriginType = t.Obj().Name()
		}

	case *types.Slice:
		if t.Elem().String() == "byte" {
			field.ProtoType = "bytes"
		} else {
			field.Repeated = true
			return s.applyFieldType(t.Elem(), field)
		}

	case *types.Pointer:
		field.Pointer = true
		if _, ok := t.Elem().(*types.Slice); ok {
			return fmt.Errorf("pointer to slice is not supported")
		}
		return s.applyFieldType(t.Elem(), field)

	case *types.Basic:
		field.ProtoType = t.String()
		field.ProtoGoType = t.String()
		field.Kind = KindBuiltin
		switch t.String() {
		case "int":
			field.ProtoType = "int64"
			field.ProtoGoType = "int64"

		case "uint":
			field.ProtoType = "uint64"
			field.ProtoGoType = "uint64"

		case "float64":
			field.ProtoType = "double"

		case "float32":
			field.ProtoType = "float"

		case "string", "bool", "uint64", "int64", "uint32", "int32":

		default:
			return fmt.Errorf("unsupported basic type %s", t)

		}

	default:
		return fmt.Errorf("unsupported type %s (%T)", t, t)
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

func implements(t types.Type, i *types.Interface) bool {
	return types.Implements(t, i) || types.Implements(types.NewPointer(t), i)
}

// Returns an iterator over the fields of a struct, and a function that reports any error that occurred.
func iterFields(n *types.Named) (iter.Seq2[*types.Var, pbTag], func() error) {
	var err error
	return func(yield func(*types.Var, pbTag) bool) {
		t, ok := n.Underlying().(*types.Struct)
		if !ok {
			err = fmt.Errorf("expected struct, got %T", n.Underlying())
			return
		}
		for i := range t.NumFields() {
			rf := t.Field(i)
			pb := reflect.StructTag(t.Tag(i)).Get("protobuf")
			if pb == "-" {
				continue
			} else if pb == "" {
				err = genErrorf(n.Obj().Pos(), "%s: missing protobuf tag", rf.Name())
				return
			}
			var pbt pbTag
			pbt, err = parsePBTag(pb)
			if err != nil {
				return
			}
			if !yield(rf, pbt) {
				return
			}
		}
	}, func() error { return err }
}

type pbTag struct {
	ID       int
	SumType  string
	Optional bool
}

func parsePBTag(tag string) (pbTag, error) {
	parts := strings.Split(tag, ",")
	if len(parts) == 0 {
		return pbTag{}, fmt.Errorf("missing tag")
	}

	idParts := strings.Split(parts[0], " ")
	sumType := ""
	if len(idParts) > 1 {
		sumType = idParts[1]
	}

	id, err := strconv.Atoi(idParts[0])
	if err != nil {
		return pbTag{}, fmt.Errorf("invalid id: %w", err)
	}
	out := pbTag{ID: id, SumType: sumType}
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

func loadInterface(pkg, symbol string) *types.Interface {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedImports | packages.NeedSyntax |
			packages.NeedFiles | packages.NeedName,
	}, pkg)
	if err != nil {
		panic(err)
	}
	for _, pkg := range pkgs {
		for _, name := range pkg.TypesInfo.Defs {
			if t, ok := name.(*types.TypeName); ok {
				if t.Name() == symbol {
					return t.Type().Underlying().(*types.Interface) //nolint:forcetypeassert
				}
			}
		}
	}
	panic("could not find " + pkg + "." + symbol)
}
