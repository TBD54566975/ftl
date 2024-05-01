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

	Comments []string `parser:"@Comment*" protobuf:"3"`
	Type     string   `parser:"@'postgres'" protobuf:"4"`
	Name     string   `parser:"'database' @Ident" protobuf:"2"`
}

var _ Decl = (*Database)(nil)
var _ Symbol = (*Database)(nil)

func (d *Database) Position() Position     { return d.Pos }
func (*Database) schemaDecl()              {}
func (*Database) schemaSymbol()            {}
func (d *Database) schemaChildren() []Node { return nil }
func (d *Database) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, encodeComments(d.Comments))
	fmt.Fprintf(w, "%s database %s", d.Type, d.Name)
	return w.String()
}

func (d *Database) ToProto() proto.Message {
	return &schemapb.Database{
		Pos:      posToProto(d.Pos),
		Name:     d.Name,
		Type:     d.Type,
		Comments: d.Comments,
	}
}

func (d *Database) GetName() string  { return d.Name }
func (d *Database) IsExported() bool { return false }

func DatabaseFromProto(s *schemapb.Database) *Database {
	return &Database{
		Pos:      posFromProto(s.Pos),
		Name:     s.Name,
		Comments: s.Comments,
		Type:     s.Type,
	}
}
