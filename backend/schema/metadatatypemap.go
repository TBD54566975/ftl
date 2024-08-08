package schema

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type MetadataTypeMap struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Runtime    string `parser:"'+' 'typemap' @('go' | 'kotlin' | 'java')" protobuf:"2"`
	NativeName string `parser:"@String" protobuf:"3"`
}

var _ Metadata = (*MetadataTypeMap)(nil)

func (m *MetadataTypeMap) Position() Position { return m.Pos }

func (m *MetadataTypeMap) String() string {
	return fmt.Sprintf("+typemap %s %q", m.Runtime, m.NativeName)
}

func (m *MetadataTypeMap) ToProto() protoreflect.ProtoMessage {
	return &schemapb.MetadataTypeMap{
		Pos:        posToProto(m.Pos),
		Runtime:    m.Runtime,
		NativeName: m.NativeName,
	}
}

func (m *MetadataTypeMap) schemaChildren() []Node { return nil }
func (m *MetadataTypeMap) schemaMetadata()        {}
