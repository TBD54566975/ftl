package schema

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

var _ Value = (*IntValue)(nil)

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
	return fmt.Sprintf("%d", i.Value)
}

func (*IntValue) schemaValueType() Type { return &Int{} }
