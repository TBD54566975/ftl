package schema

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
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

func (m *MetadataSQLMigration) ToProto() proto.Message {
	return &schemapb.MetadataSQLMigration{
		Pos:    posToProto(m.Pos),
		Digest: m.Digest,
	}
}
