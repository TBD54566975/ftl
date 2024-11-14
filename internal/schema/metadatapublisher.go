package schema

import (
	"fmt"
	"strings"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"google.golang.org/protobuf/proto"
)

//protobuf:12,optional
type MetadataPublisher struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Topics []*Ref `parser:"'+' 'publish' @@ (',' @@)*" protobuf:"2"`
}

var _ Metadata = (*MetadataPublisher)(nil)

func (m *MetadataPublisher) Position() Position { return m.Pos }
func (m *MetadataPublisher) String() string {
	out := &strings.Builder{}
	fmt.Fprint(out, "+publish ")
	w := 6
	for i, topic := range m.Topics {
		if i > 0 {
			fmt.Fprint(out, ", ")
			w += 2
		}
		str := topic.String()
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

func (m *MetadataPublisher) schemaChildren() []Node {
	out := make([]Node, 0, len(m.Topics))
	for _, ref := range m.Topics {
		out = append(out, ref)
	}
	return out
}
func (*MetadataPublisher) schemaMetadata() {}

func (m *MetadataPublisher) ToProto() proto.Message {
	return &schemapb.MetadataPublisher{
		Pos:    posToProto(m.Pos),
		Topics: nodeListToProto[*schemapb.Ref](m.Topics),
	}
}
