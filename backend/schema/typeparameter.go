package schema

import (
	"google.golang.org/protobuf/reflect/protoreflect"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type TypeParameter struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Name string `parser:"@Ident" protobuf:"2"`
}

var _ Symbol = (*TypeParameter)(nil)

func (t *TypeParameter) Position() Position { return t.Pos }
func (t *TypeParameter) String() string     { return t.Name }
func (t *TypeParameter) ToProto() protoreflect.ProtoMessage {
	return &schemapb.TypeParameter{Pos: posToProto(t.Pos), Name: t.Name}
}
func (t *TypeParameter) schemaChildren() []Node { return nil }
func (t *TypeParameter) schemaSymbol()          {}
func (t *TypeParameter) GetName() string        { return t.Name }

func typeParametersToSchema(s []*schemapb.TypeParameter) []*TypeParameter {
	var out []*TypeParameter
	for _, n := range s {
		out = append(out, &TypeParameter{
			Pos:  posFromProto(n.Pos),
			Name: n.Name,
		})
	}
	return out
}
