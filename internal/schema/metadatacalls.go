package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

// MetadataCalls represents a metadata block with a list of calls.
//
//protobuf:1,optional
type MetadataCalls struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Calls []*Ref `parser:"'+' 'calls' @@ (',' @@)*" protobuf:"2"`
}

var _ Metadata = (*MetadataCalls)(nil)

func (m *MetadataCalls) Position() Position { return m.Pos }
func (m *MetadataCalls) String() string {
	out := &strings.Builder{}
	fmt.Fprint(out, "+calls ")
	w := 6
	for i, call := range m.Calls {
		if i > 0 {
			fmt.Fprint(out, ", ")
			w += 2
		}
		str := call.String()
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

func (m *MetadataCalls) schemaChildren() []Node {
	out := make([]Node, 0, len(m.Calls))
	for _, ref := range m.Calls {
		out = append(out, ref)
	}
	return out
}
func (*MetadataCalls) schemaMetadata() {}

func (m *MetadataCalls) ToProto() proto.Message {
	return &schemapb.MetadataCalls{
		Pos:   posToProto(m.Pos),
		Calls: nodeListToProto[*schemapb.Ref](m.Calls),
	}
}
