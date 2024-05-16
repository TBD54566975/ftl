package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type Config struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" protobuf:"2"`
	Name     string   `parser:"'config' @Ident" protobuf:"3"`
	Type     Type     `parser:"@@" protobuf:"4"`
}

var _ Decl = (*Config)(nil)
var _ Symbol = (*Config)(nil)

func (s *Config) GetName() string    { return s.Name }
func (s *Config) IsExported() bool   { return false }
func (s *Config) Position() Position { return s.Pos }
func (s *Config) String() string {
	w := &strings.Builder{}

	fmt.Fprint(w, EncodeComments(s.Comments))
	fmt.Fprintf(w, "config %s %s", s.Name, s.Type)

	return w.String()
}

func (s *Config) ToProto() protoreflect.ProtoMessage {
	return &schemapb.Config{
		Pos:      posToProto(s.Pos),
		Comments: s.Comments,
		Name:     s.Name,
		Type:     typeToProto(s.Type),
	}
}

func (s *Config) schemaChildren() []Node { return []Node{s.Type} }
func (s *Config) schemaDecl()            {}
func (s *Config) schemaSymbol()          {}

func ConfigFromProto(p *schemapb.Config) *Config {
	return &Config{
		Pos:      posFromProto(p.Pos),
		Name:     p.Name,
		Comments: p.Comments,
		Type:     typeToSchema(p.Type),
	}
}
