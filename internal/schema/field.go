package schema

import (
	"fmt"
	"strings"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type Field struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string   `parser:"@Comment*" protobuf:"3"`
	Name     string     `parser:"@Ident" protobuf:"2"`
	Type     Type       `parser:"@@" protobuf:"4"`
	Metadata []Metadata `parser:"@@*" protobuf:"5"`
}

var _ Node = (*Field)(nil)

func (f *Field) Position() Position { return f.Pos }
func (f *Field) schemaChildren() []Node {
	out := []Node{}
	out = append(out, f.Type)
	for _, md := range f.Metadata {
		out = append(out, md)
	}
	return out
}
func (f *Field) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(f.Comments))
	fmt.Fprintf(w, "%s %s", f.Name, f.Type.String())
	for _, md := range f.Metadata {
		fmt.Fprintf(w, " %s", md.String())
	}
	return w.String()
}

func (f *Field) ToProto() proto.Message {
	return &schemapb.Field{
		Pos:      posToProto(f.Pos),
		Name:     f.Name,
		Type:     TypeToProto(f.Type),
		Comments: f.Comments,
		Metadata: metadataListToProto(f.Metadata),
	}
}

// Alias returns the alias for the given kind.
func (f *Field) Alias(kind AliasKind) optional.Option[string] {
	for _, md := range f.Metadata {
		if a, ok := md.(*MetadataAlias); ok && a.Kind == kind {
			return optional.Some(a.Alias)
		}
	}
	return optional.None[string]()
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
		Pos:      PosFromProto(s.Pos),
		Name:     s.Name,
		Comments: s.Comments,
		Type:     TypeFromProto(s.Type),
		Metadata: metadataListToSchema(s.Metadata),
	}
}
