package schema

import (
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

var _ Value = (*TypeValue)(nil)

//protobuf:3
type TypeValue struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Value Type `parser:"@@" protobuf:"2"`
}

func (t *TypeValue) ToProto() proto.Message {
	return &schemapb.TypeValue{
		Pos:   posToProto(t.Pos),
		Value: TypeToProto(t.Value),
	}
}

func (t *TypeValue) Position() Position { return t.Pos }

func (t *TypeValue) schemaChildren() []Node { return []Node{t.Value} }

func (t *TypeValue) String() string {
	return t.Value.String()
}

func (t *TypeValue) GetValue() any { return t.Value.String() }

func (t *TypeValue) schemaValueType() Type { return t.Value }
