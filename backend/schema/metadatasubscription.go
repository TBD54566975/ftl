package schema

import (
	"fmt"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"google.golang.org/protobuf/proto"
)

type MetadataSubscriber struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Name string `parser:"'+' 'subscribe' @Ident" protobuf:"2"`
}

var _ Metadata = (*MetadataRetry)(nil)

func (*MetadataSubscriber) schemaMetadata()          {}
func (m *MetadataSubscriber) schemaChildren() []Node { return nil }
func (m *MetadataSubscriber) Position() Position     { return m.Pos }
func (m *MetadataSubscriber) String() string {
	return fmt.Sprintf("+subscribe %v", m.Name)
}

func (m *MetadataSubscriber) ToProto() proto.Message {
	return &schemapb.MetadataSubscriber{
		Pos:  posToProto(m.Pos),
		Name: m.Name,
	}
}
