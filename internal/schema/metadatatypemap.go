package schema

import (
	"fmt"
)

//protobuf:8
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

func (m *MetadataTypeMap) schemaChildren() []Node { return nil }
func (m *MetadataTypeMap) schemaMetadata()        {}
