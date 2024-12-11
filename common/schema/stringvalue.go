package schema

import (
	"fmt"
)

//protobuf:1
type StringValue struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Value string `parser:"@String" protobuf:"2"`
}

var _ Value = (*StringValue)(nil)

func (s *StringValue) Position() Position { return s.Pos }

func (s *StringValue) schemaChildren() []Node { return nil }

func (s *StringValue) String() string {
	return fmt.Sprintf("\"%s\"", s.Value)
}

func (s *StringValue) GetValue() any { return s.Value }

func (*StringValue) schemaValueType() Type { return &String{} }
