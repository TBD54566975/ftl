package schema

import (
	"fmt"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"google.golang.org/protobuf/proto"
)

type StringValue struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Value string `parser:"@String" protobuf:"2"`
}

var _ Value = (*StringValue)(nil)

func (s *StringValue) ToProto() proto.Message {
	return &schemapb.StringValue{
		Pos:   posToProto(s.Pos),
		Value: s.Value,
	}
}

func (s *StringValue) Position() Position { return s.Pos }

func (s *StringValue) schemaChildren() []Node { return nil }

func (s *StringValue) String() string {
	return fmt.Sprintf("\"%s\"", s.Value)
}

func (*StringValue) schemaValueType() Type { return &String{} }
