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

	Comments []string   `parser:"@Comment*" protobuf:"5"`
	Name     string     `parser:"'data' @Ident '{'" protobuf:"2"`
	Fields   []*Field   `parser:"@@* '}'" protobuf:"3"`
	Metadata []Metadata `parser:"@@*" protobuf:"4"`
}

var _ Decl = (*Data)(nil)

// schemaDecl implements Decl
func (*Data) schemaDecl() {}
func (d *Data) schemaChildren() []Node {
	children := make([]Node, 0, len(d.Fields)+len(d.Metadata))
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
	fmt.Fprintf(w, "data %s {\n", d.Name)
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
