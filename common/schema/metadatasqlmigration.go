package schema

import (
	"fmt"
)

//protobuf:13
type MetadataSQLMigration struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Digest string `parser:"'+' 'migration' Whitespace 'sha256' ':' @SHA256" protobuf:"2"`
}

var _ Metadata = (*MetadataSQLMigration)(nil)

func (*MetadataSQLMigration) schemaMetadata()          {}
func (m *MetadataSQLMigration) schemaChildren() []Node { return nil }
func (m *MetadataSQLMigration) Position() Position     { return m.Pos }
func (m *MetadataSQLMigration) String() string {
	return fmt.Sprintf("+migration sha256:%v", m.Digest)
}
