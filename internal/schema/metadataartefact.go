package schema

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
)

//protobuf:14
type MetadataArtefact struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Path       string `parser:"'+' 'artefact' Whitespace @String" protobuf:"2"`
	Digest     string `parser:"@String" protobuf:"3"`
	Executable bool   `parser:"@'executable'?" protobuf:"4"`
}

var _ Metadata = (*MetadataArtefact)(nil)

func (m *MetadataArtefact) Position() Position { return m.Pos }

func (m *MetadataArtefact) String() string {
	return fmt.Sprintf("+artefact %q %q %t", m.Path, m.Digest, m.Executable)
}

func (m *MetadataArtefact) ToProto() protoreflect.ProtoMessage {
	return &schemapb.MetadataArtefact{
		Pos:        posToProto(m.Pos),
		Executable: m.Executable,
		Path:       m.Path,
		Digest:     m.Digest,
	}
}

func (m *MetadataArtefact) schemaChildren() []Node { return nil }
func (m *MetadataArtefact) schemaMetadata()        {}
