package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

type Field struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" protobuf:"3"`
	Name     string   `parser:"@Ident" protobuf:"2"`
	Type     Type     `parser:"@@" protobuf:"4"`
}

var _ Node = (*Field)(nil)

func (f *Field) schemaChildren() []Node { return []Node{f.Type} }
func (f *Field) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, encodeComments(f.Comments))
	fmt.Fprintf(w, "%s %s", f.Name, f.Type.String())
	return w.String()
}

func (f *Field) ToProto() proto.Message {
	return &schemapb.Field{
		Pos:      posToProto(f.Pos),
		Name:     f.Name,
		Type:     typeToProto(f.Type),
		Comments: f.Comments,
	}
}

func fieldListToSchema(s []*schemapb.Field) []*Field {
	var out []*Field
	for _, n := range s {
		out = append(out, fieldToSchema(n))
	}
	return out
}

func fieldToSchema(s *schemapb.Field) *Field {
	return &Field{
		Pos:      posFromProto(s.Pos),
		Name:     s.Name,
		Comments: s.Comments,
		Type:     typeToSchema(s.Type),
	}
}
