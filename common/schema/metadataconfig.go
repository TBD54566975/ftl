package schema

import (
	"fmt"
	"strings"
)

// MetadataConfig represents a metadata block with a list of config items that are used.
//
//protobuf:10,optional
type MetadataConfig struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Config []*Ref `parser:"'+' 'config' @@ (',' @@)*" protobuf:"2"`
}

var _ Metadata = (*MetadataConfig)(nil)

func (m *MetadataConfig) Position() Position { return m.Pos }
func (m *MetadataConfig) String() string {
	out := &strings.Builder{}
	fmt.Fprint(out, "+config ")
	w := 6
	for i, config := range m.Config {
		if i > 0 {
			fmt.Fprint(out, ", ")
			w += 2
		}
		str := config.String()
		if w+len(str) > 70 {
			w = 6
			fmt.Fprint(out, "\n      ")
		}
		w += len(str)
		fmt.Fprint(out, str)
	}
	fmt.Fprint(out)
	return out.String()
}

func (m *MetadataConfig) schemaChildren() []Node {
	out := make([]Node, 0, len(m.Config))
	for _, ref := range m.Config {
		out = append(out, ref)
	}
	return out
}
func (*MetadataConfig) schemaMetadata() {}
