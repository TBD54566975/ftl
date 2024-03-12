package schema

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type Config struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Name string `parser:"'config' @Ident" protobuf:"2"`
	Type Type   `parser:"@@" protobuf:"3"`
}

var _ Decl = (*Config)(nil)

func (s *Config) GetName() string    { return s.Name }
func (s *Config) Position() Position { return s.Pos }
func (s *Config) String() string     { return fmt.Sprintf("config %s %s", s.Name, s.Type) }

func (s *Config) ToProto() protoreflect.ProtoMessage {
	return &schemapb.Config{
		Pos:  posToProto(s.Pos),
		Name: s.Name,
		Type: typeToProto(s.Type),
	}
}

func (s *Config) schemaChildren() []Node { return []Node{s.Type} }

func (s *Config) schemaDecl() {}

func ConfigFromProto(p *schemapb.Config) *Config {
	return &Config{
		Pos:  posFromProto(p.Pos),
		Name: p.Name,
		Type: typeToSchema(p.Type),
	}
}
