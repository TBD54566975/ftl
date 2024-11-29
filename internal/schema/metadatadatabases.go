package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
)

//protobuf:4
type MetadataDatabases struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Calls []*Ref `parser:"'+' 'database' 'calls' @@ (',' @@)*" protobuf:"2"`
}

var _ Metadata = (*MetadataDatabases)(nil)

func (m *MetadataDatabases) Position() Position { return m.Pos }
func (m *MetadataDatabases) String() string {
	out := &strings.Builder{}
	fmt.Fprint(out, "+database calls ")
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

func (m *MetadataDatabases) schemaChildren() []Node {
	out := make([]Node, 0, len(m.Calls))
	for _, ref := range m.Calls {
		out = append(out, ref)
	}
	return out
}
func (*MetadataDatabases) schemaMetadata() {}

func (m *MetadataDatabases) ToProto() proto.Message {
	return &schemapb.MetadataDatabases{
		Pos:   posToProto(m.Pos),
		Calls: nodeListToProto[*schemapb.Ref](m.Calls),
	}
}
