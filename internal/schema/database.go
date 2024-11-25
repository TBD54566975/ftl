package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

const PostgresDatabaseType = "postgres"
const MySQLDatabaseType = "mysql"

//protobuf:3
type Database struct {
	Pos     Position        `parser:"" protobuf:"1,optional"`
	Runtime DatabaseRuntime `parser:"" protobuf:"31634,optional"`

	Comments []string   `parser:"@Comment*" protobuf:"2"`
	Type     string     `parser:"'database' @('postgres'|'mysql')" protobuf:"4"`
	Name     string     `parser:"@Ident" protobuf:"3"`
	Metadata []Metadata `parser:"@@*" protobuf:"5"`
}

var _ Decl = (*Database)(nil)
var _ Symbol = (*Database)(nil)

func (d *Database) Position() Position { return d.Pos }
func (*Database) schemaDecl()          {}
func (*Database) schemaSymbol()        {}
func (d *Database) schemaChildren() []Node {
	children := []Node{}
	for _, c := range d.Metadata {
		children = append(children, c)
	}
	return children
}
func (d *Database) Redact() { d.Runtime = nil }
func (d *Database) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(d.Comments))
	fmt.Fprintf(w, "database %s %s", d.Type, d.Name)
	fmt.Fprint(w, indent(encodeMetadata(d.Metadata)))
	return w.String()
}

func (d *Database) ToProto() proto.Message {
	var runtime *schemapb.DatabaseRuntime
	if d.Runtime != nil {
		r, ok := d.Runtime.ToProto().(*schemapb.DatabaseRuntime)
		if !ok {
			panic(fmt.Sprintf("unknown database runtime type: %T", d.Runtime))
		}
		runtime = r
	}

	return &schemapb.Database{
		Pos:      posToProto(d.Pos),
		Comments: d.Comments,
		Name:     d.Name,
		Type:     d.Type,
		Runtime:  runtime,
		Metadata: metadataListToProto(d.Metadata),
	}
}

func (d *Database) GetName() string  { return d.Name }
func (d *Database) IsExported() bool { return false }

func DatabaseFromProto(s *schemapb.Database) *Database {
	return &Database{
		Pos:      PosFromProto(s.Pos),
		Comments: s.Comments,
		Name:     s.Name,
		Type:     s.Type,
		Runtime:  DatabaseRuntimeFromProto(s.Runtime),
		Metadata: metadataListToSchema(s.Metadata),
	}
}

type DatabaseRuntime interface {
	Symbol
	databaseRuntime()
}

//protobuf:1
type DSNDatabaseRuntime struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	DSN string `parser:"" protobuf:"2"`
}

var _ DatabaseRuntime = (*DSNDatabaseRuntime)(nil)

func (d *DSNDatabaseRuntime) Position() Position { return d.Pos }
func (d *DSNDatabaseRuntime) databaseRuntime()   {}
func (d *DSNDatabaseRuntime) String() string     { return d.DSN }
func (d *DSNDatabaseRuntime) ToProto() protoreflect.ProtoMessage {
	if d == nil {
		return nil
	}
	return &schemapb.DatabaseRuntime{
		Value: &schemapb.DatabaseRuntime_DsnDatabaseRuntime{
			DsnDatabaseRuntime: &schemapb.DSNDatabaseRuntime{Dsn: d.DSN},
		},
	}
}

func (d *DSNDatabaseRuntime) schemaChildren() []Node { return nil }
func (d *DSNDatabaseRuntime) schemaSymbol()          {}

func DatabaseRuntimeFromProto(s *schemapb.DatabaseRuntime) DatabaseRuntime {
	if s == nil {
		return nil
	}
	switch s := s.Value.(type) {
	case *schemapb.DatabaseRuntime_DsnDatabaseRuntime:
		return &DSNDatabaseRuntime{DSN: s.DsnDatabaseRuntime.Dsn}
	default:
		panic(fmt.Sprintf("unknown database runtime type: %T", s))
	}
}
