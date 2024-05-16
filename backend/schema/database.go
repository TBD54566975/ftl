package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

const PostgresDatabaseType = "postgres"

type Database struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Comments []string `parser:"@Comment*" protobuf:"2"`
	Name     string   `parser:"@Ident" protobuf:"3"`
	Type     string   `parser:"'database' @'postgres'" protobuf:"4"`
}

var _ Decl = (*Database)(nil)
var _ Symbol = (*Database)(nil)

func (d *Database) Position() Position     { return d.Pos }
func (*Database) schemaDecl()              {}
func (*Database) schemaSymbol()            {}
func (d *Database) schemaChildren() []Node { return nil }
func (d *Database) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(d.Comments))
	fmt.Fprintf(w, "database %s %s", d.Type, d.Name)
	return w.String()
}

func (d *Database) ToProto() proto.Message {
	return &schemapb.Database{
		Pos:      posToProto(d.Pos),
		Comments: d.Comments,
		Name:     d.Name,
		Type:     d.Type,
	}
}

func (d *Database) GetName() string  { return d.Name }
func (d *Database) IsExported() bool { return false }

func DatabaseFromProto(s *schemapb.Database) *Database {
	return &Database{
		Pos:      posFromProto(s.Pos),
		Comments: s.Comments,
		Name:     s.Name,
		Type:     s.Type,
	}
}
