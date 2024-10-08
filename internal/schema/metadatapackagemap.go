package schema

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

//protobuf:12
type MetadataPackageMap struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Runtime string `parser:"'+' 'packagemap' @('go' | 'kotlin' | 'java')" protobuf:"2"`
	Package string `parser:"@String" protobuf:"3"`
}

var _ Metadata = (*MetadataPackageMap)(nil)

func (m *MetadataPackageMap) Position() Position { return m.Pos }

func (m *MetadataPackageMap) String() string {
	return fmt.Sprintf("+packagemap %s %q", m.Runtime, m.Package)
}

func (m *MetadataPackageMap) ToProto() protoreflect.ProtoMessage {
	return &schemapb.MetadataPackageMap{
		Pos:     posToProto(m.Pos),
		Runtime: m.Runtime,
		Package: m.Package,
	}
}

func (m *MetadataPackageMap) schemaChildren() []Node { return nil }
func (m *MetadataPackageMap) schemaMetadata()        {}
