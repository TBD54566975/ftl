package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type TypeAlias struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string   `parser:"@Comment*" protobuf:"2"`
	Export   bool       `parser:"@'export'?" protobuf:"3"`
	Name     string     `parser:"'typealias' @Ident" protobuf:"4"`
	Type     Type       `parser:"@@" protobuf:"5"`
	Metadata []Metadata `parser:"@@*" protobuf:"6"`
}

var _ Decl = (*TypeAlias)(nil)
var _ Symbol = (*TypeAlias)(nil)

func (t *TypeAlias) Position() Position { return t.Pos }

func (t *TypeAlias) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(t.Comments))
	if t.Export {
		fmt.Fprint(w, "export ")
	}
	fmt.Fprintf(w, "typealias %s %s", t.Name, t.Type)
	fmt.Fprint(w, indent(encodeMetadata(t.Metadata)))
	return w.String()
}
func (*TypeAlias) schemaDecl()   {}
func (*TypeAlias) schemaSymbol() {}
func (t *TypeAlias) schemaChildren() []Node {
	children := make([]Node, 0, len(t.Metadata)+1)
	for _, m := range t.Metadata {
		children = append(children, m)
	}
	if t.Type != nil {
		children = append(children, t.Type)
	}
	return children
}
func (t *TypeAlias) ToProto() proto.Message {
	return &schemapb.TypeAlias{
		Pos:      posToProto(t.Pos),
		Comments: t.Comments,
		Name:     t.Name,
		Export:   t.Export,
		Type:     TypeToProto(t.Type),
		Metadata: metadataListToProto(t.Metadata),
	}
}

func (t *TypeAlias) GetName() string  { return t.Name }
func (t *TypeAlias) IsExported() bool { return t.Export }

func TypeAliasFromProto(s *schemapb.TypeAlias) *TypeAlias {
	return &TypeAlias{
		Pos:      posFromProto(s.Pos),
		Name:     s.Name,
		Export:   s.Export,
		Comments: s.Comments,
		Type:     TypeFromProto(s.Type),
		Metadata: metadataListToSchema(s.Metadata),
	}
}
