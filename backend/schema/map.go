package schema

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type Map struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Key   Type `parser:"'{' @@" protobuf:"2"`
	Value Type `parser:"':' @@ '}'" protobuf:"3"`
}

var _ Type = (*Map)(nil)
var _ Symbol = (*Map)(nil)

func (m *Map) Equal(other Type) bool {
	o, ok := other.(*Map)
	if !ok {
		return false
	}
	return m.Key.Equal(o.Key) && m.Value.Equal(o.Value)
}
func (m *Map) Position() Position     { return m.Pos }
func (m *Map) schemaChildren() []Node { return []Node{m.Key, m.Value} }
func (*Map) schemaType()              {}
func (*Map) schemaSymbol()            {}
func (m *Map) String() string         { return fmt.Sprintf("{%s: %s}", m.Key.String(), m.Value.String()) }

func (m *Map) ToProto() proto.Message {
	return &schemapb.Map{
		Key:   TypeToProto(m.Key),
		Value: TypeToProto(m.Value),
	}
}

func mapToSchema(s *schemapb.Map) *Map {
	return &Map{
		Pos:   posFromProto(s.Pos),
		Key:   TypeFromProto(s.Key),
		Value: TypeFromProto(s.Value),
	}
}
