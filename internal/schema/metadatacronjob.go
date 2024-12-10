package schema

import (
	"fmt"
)

//protobuf:3
type MetadataCronJob struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Cron string `parser:"'+' 'cron' Whitespace @(' ' | ~EOL)+" protobuf:"2"`
}

var _ Metadata = (*MetadataCronJob)(nil)

func (m *MetadataCronJob) Position() Position { return m.Pos }
func (m *MetadataCronJob) String() string {
	return fmt.Sprintf("+cron %s", m.Cron)
}

func (m *MetadataCronJob) schemaChildren() []Node {
	return nil
}

func (*MetadataCronJob) schemaMetadata() {}
