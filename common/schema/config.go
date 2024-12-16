package schema

import (
	"fmt"
	"strings"

	schemapb "github.com/block/ftl/common/protos/xyz/block/ftl/schema/v1"
)

//protobuf:6
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

func (s *Config) schemaChildren() []Node { return []Node{s.Type} }
func (s *Config) schemaDecl()            {}
func (s *Config) schemaSymbol()          {}

func ConfigFromProto(p *schemapb.Config) *Config {
	return &Config{
		Pos:      PosFromProto(p.Pos),
		Comments: p.Comments,
		Name:     p.Name,
		Type:     TypeFromProto(p.Type),
	}
}
