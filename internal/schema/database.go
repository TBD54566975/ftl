package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
)

const PostgresDatabaseType = "postgres"
const MySQLDatabaseType = "mysql"

//protobuf:3
type Database struct {
	Pos     Position         `parser:"" protobuf:"1,optional"`
	Runtime *DatabaseRuntime `parser:"" protobuf:"31634,optional"`

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

type DatabaseRuntime struct {
	ReadConnector  DatabaseConnector `parser:"" protobuf:"1"`
	WriteConnector DatabaseConnector `parser:"" protobuf:"2"`
}

var _ Runtime = (*DatabaseRuntime)(nil)
var _ Symbol = (*DatabaseRuntime)(nil)

func (d *DatabaseRuntime) runtime()           {}
func (d *DatabaseRuntime) Position() Position { return d.ReadConnector.Position() }
func (d *DatabaseRuntime) schemaSymbol()      {}
func (d *DatabaseRuntime) String() string {
	return fmt.Sprintf("read: %s, write: %s", d.ReadConnector, d.WriteConnector)
}
func (d *DatabaseRuntime) ToProto() protoreflect.ProtoMessage {
	rc, ok := d.ReadConnector.ToProto().(*schemapb.DatabaseConnector)
	if !ok {
		panic(fmt.Sprintf("unknown database connector type: %T", d.ReadConnector))
	}
	wc, ok := d.WriteConnector.ToProto().(*schemapb.DatabaseConnector)
	if !ok {
		panic(fmt.Sprintf("unknown database connector type: %T", d.WriteConnector))
	}

	return &schemapb.DatabaseRuntime{
		ReadConnector:  rc,
		WriteConnector: wc,
	}
}
func (d *DatabaseRuntime) schemaChildren() []Node { return []Node{d.ReadConnector, d.WriteConnector} }

type DatabaseConnector interface {
	Node

	databaseConnector()
}

//protobuf:1
type DSNDatabaseConnector struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	DSN string `parser:"" protobuf:"2"`
}

var _ DatabaseConnector = (*DSNDatabaseConnector)(nil)

func (d *DSNDatabaseConnector) Position() Position { return d.Pos }
func (d *DSNDatabaseConnector) databaseConnector() {}
func (d *DSNDatabaseConnector) String() string     { return d.DSN }
func (d *DSNDatabaseConnector) ToProto() protoreflect.ProtoMessage {
	if d == nil {
		return nil
	}
	return &schemapb.DatabaseConnector{
		Value: &schemapb.DatabaseConnector_DsnDatabaseConnector{
			DsnDatabaseConnector: &schemapb.DSNDatabaseConnector{Dsn: d.DSN},
		},
	}
}

func (d *DSNDatabaseConnector) schemaChildren() []Node { return nil }

//protobuf:2
type AWSIAMAuthDatabaseConnector struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Username string `parser:"" protobuf:"2"`
	Endpoint string `parser:"" protobuf:"3"`
	Database string `parser:"" protobuf:"4"`
}

var _ DatabaseConnector = (*AWSIAMAuthDatabaseConnector)(nil)

func (d *AWSIAMAuthDatabaseConnector) Position() Position { return d.Pos }
func (d *AWSIAMAuthDatabaseConnector) databaseConnector() {}
func (d *AWSIAMAuthDatabaseConnector) String() string {
	return fmt.Sprintf("%s@%s/%s", d.Username, d.Endpoint, d.Database)
}
func (d *AWSIAMAuthDatabaseConnector) ToProto() protoreflect.ProtoMessage {
	if d == nil {
		return nil
	}
	return &schemapb.DatabaseConnector{
		Value: &schemapb.DatabaseConnector_AwsiamAuthDatabaseConnector{
			AwsiamAuthDatabaseConnector: &schemapb.AWSIAMAuthDatabaseConnector{
				Username: d.Username,
				Endpoint: d.Endpoint,
				Database: d.Database,
			},
		},
	}
}

func (d *AWSIAMAuthDatabaseConnector) schemaChildren() []Node { return nil }

func DatabaseRuntimeFromProto(s *schemapb.DatabaseRuntime) *DatabaseRuntime {
	if s == nil {
		return nil
	}
	return &DatabaseRuntime{
		ReadConnector:  DatabaseConnectorFromProto(s.ReadConnector),
		WriteConnector: DatabaseConnectorFromProto(s.WriteConnector),
	}
}

func DatabaseConnectorFromProto(s *schemapb.DatabaseConnector) DatabaseConnector {
	switch s := s.Value.(type) {
	case *schemapb.DatabaseConnector_DsnDatabaseConnector:
		return &DSNDatabaseConnector{DSN: s.DsnDatabaseConnector.Dsn}
	case *schemapb.DatabaseConnector_AwsiamAuthDatabaseConnector:
		return &AWSIAMAuthDatabaseConnector{
			Username: s.AwsiamAuthDatabaseConnector.Username,
			Endpoint: s.AwsiamAuthDatabaseConnector.Endpoint,
			Database: s.AwsiamAuthDatabaseConnector.Database,
		}
	default:
		panic(fmt.Sprintf("unknown database connector type: %T", s))
	}
}
