package schema

import (
	"fmt"
	"strings"

	schemapb "github.com/TBD54566975/ftl/common/protos/xyz/block/ftl/schema/v1"
	"github.com/TBD54566975/ftl/common/slices"
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
var _ Provisioned = (*Database)(nil)

func (d *Database) Position() Position { return d.Pos }
func (*Database) schemaDecl()          {}
func (*Database) schemaSymbol()        {}
func (d *Database) provisioned()       {}
func (d *Database) schemaChildren() []Node {
	children := []Node{}
	for _, c := range d.Metadata {
		children = append(children, c)
	}
	return children
}
func (d *Database) String() string {
	w := &strings.Builder{}
	fmt.Fprint(w, EncodeComments(d.Comments))
	fmt.Fprintf(w, "database %s %s", d.Type, d.Name)
	fmt.Fprint(w, indent(encodeMetadata(d.Metadata)))
	return w.String()
}

func (d *Database) GetName() string  { return d.Name }
func (d *Database) IsExported() bool { return false }

func (d *Database) GetProvisioned() ResourceSet {
	kind := ResourceTypeMysql
	if d.Type == PostgresDatabaseType {
		kind = ResourceTypePostgres
	}
	result := []*ProvisionedResource{{
		Kind:   kind,
		Config: &Database{Type: d.Type},
	}}

	migration, ok := slices.FindVariant[*MetadataSQLMigration](d.Metadata)
	if ok {
		result = append(result, &ProvisionedResource{
			Kind:   ResourceTypeSQLMigration,
			Config: &Database{Type: d.Type, Metadata: []Metadata{migration}},
		})
	}

	return result

}

func (d *Database) ResourceID() string {
	return d.Name
}

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