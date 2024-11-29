package schema

import (
	"strconv"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/schema/v1"
)

var _ Value = (*IntValue)(nil)

//protobuf:2
type IntValue struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Value int `parser:"@Number" protobuf:"2"`
}

func (i *IntValue) ToProto() proto.Message {
	return &schemapb.IntValue{
		Pos:   posToProto(i.Pos),
		Value: int64(i.Value),
	}
}

func (i *IntValue) Position() Position { return i.Pos }

func (i *IntValue) schemaChildren() []Node { return nil }

func (i *IntValue) String() string {
	return strconv.Itoa(i.Value)
}

func (i *IntValue) GetValue() any { return i.Value }

func (*IntValue) schemaValueType() Type { return &Int{} }
