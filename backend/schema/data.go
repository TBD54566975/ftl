package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

// A Data structure.
type Data struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments       []string         `parser:"@Comment*" protobuf:"2"`
	Name           string           `parser:"'data' @Ident" protobuf:"3"`
	TypeParameters []*TypeParameter `parser:"( '<' @@ (',' @@)* '>' )?" protobuf:"6"`
	Fields         []*Field         `parser:"'{' @@* '}'" protobuf:"4"`
	Metadata       []Metadata       `parser:"@@*" protobuf:"5"`
}

var _ Decl = (*Data)(nil)
var _ Scoped = (*Data)(nil)

func (d *Data) Scope() Scope {
	scope := Scope{}
	for _, t := range d.TypeParameters {
		scope[t.Name] = ModuleDecl{Decl: t}
	}
	return scope
}

func (d *Data) Position() Position { return d.Pos }
func (*Data) schemaDecl()          {}
func (d *Data) schemaChildren() []Node {
	children := make([]Node, 0, len(d.Fields)+len(d.Metadata))
	for _, t := range d.TypeParameters {
		children = append(children, t)
	}
	for _, f := range d.Fields {
		children = append(children, f)
	}
	for _, c := range d.Metadata {
		children = append(children, c)
	}
	return children
}

func (d *Data) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, encodeComments(d.Comments))
	typeParameters := ""
	if len(d.TypeParameters) > 0 {
		typeParameters = "<"
		for i, t := range d.TypeParameters {
			if i != 0 {
				typeParameters += ", "
			}
			typeParameters += t.String()
		}
		typeParameters += ">"
	}
	fmt.Fprintf(w, "data %s%s {\n", d.Name, typeParameters)
	for _, f := range d.Fields {
		fmt.Fprintln(w, indent(f.String()))
	}
	fmt.Fprintf(w, "}")
	fmt.Fprint(w, indent(encodeMetadata(d.Metadata)))
	return w.String()
}

func (d *Data) ToProto() proto.Message {
	return &schemapb.Data{
		Pos:      posToProto(d.Pos),
		Name:     d.Name,
		Fields:   nodeListToProto[*schemapb.Field](d.Fields),
		Comments: d.Comments,
	}
}

func DataToSchema(s *schemapb.Data) *Data {
	return &Data{
		Pos:      posFromProto(s.Pos),
		Name:     s.Name,
		Fields:   fieldListToSchema(s.Fields),
		Comments: s.Comments,
	}
}
