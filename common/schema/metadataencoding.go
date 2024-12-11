package schema

import (
	"fmt"
	"strings"
)

//protobuf:9
type MetadataEncoding struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Type    string `parser:"'+' 'encoding' @'json'" protobuf:"2"`
	Lenient bool   `parser:"@'lenient'?" protobuf:"3"`
}

var _ Metadata = (*MetadataEncoding)(nil)

func (m *MetadataEncoding) Position() Position { return m.Pos }

func (m *MetadataEncoding) String() string {
	w := &strings.Builder{}
	if m.Type == "" {
		fmt.Fprintf(w, "+encoding json")
	} else {
		fmt.Fprintf(w, "+encoding %s", m.Type)
	}
	if m.Lenient {
		fmt.Fprintf(w, " lenient")
	}
	return w.String()
}

func (m *MetadataEncoding) schemaChildren() []Node { return nil }
func (m *MetadataEncoding) schemaMetadata()        {}
