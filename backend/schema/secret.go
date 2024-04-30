package schema

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type Secret struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Name string `parser:"'secret' @Ident" protobuf:"2"`
	Type Type   `parser:"@@" protobuf:"3"`
}

var _ Decl = (*Secret)(nil)
var _ Symbol = (*Secret)(nil)

func (s *Secret) GetName() string    { return s.Name }
func (s *Secret) IsExported() bool   { return false }
func (s *Secret) Position() Position { return s.Pos }
func (s *Secret) String() string     { return fmt.Sprintf("secret %s %s", s.Name, s.Type) }

func (s *Secret) ToProto() protoreflect.ProtoMessage {
	return &schemapb.Secret{
		Pos:  posToProto(s.Pos),
		Name: s.Name,
		Type: typeToProto(s.Type),
	}
}

func (s *Secret) schemaChildren() []Node { return []Node{s.Type} }

func (s *Secret) schemaDecl()   {}
func (s *Secret) schemaSymbol() {}

func SecretFromProto(p *schemapb.Secret) *Secret {
	return &Secret{
		Pos:  posFromProto(p.Pos),
		Name: p.Name,
		Type: typeToSchema(p.Type),
	}
}
