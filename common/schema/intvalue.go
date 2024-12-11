package schema

import (
	"strconv"
)

var _ Value = (*IntValue)(nil)

//protobuf:2
type IntValue struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Value int `parser:"@Number" protobuf:"2"`
}

func (i *IntValue) Position() Position { return i.Pos }

func (i *IntValue) schemaChildren() []Node { return nil }

func (i *IntValue) String() string {
	return strconv.Itoa(i.Value)
}

func (i *IntValue) GetValue() any { return i.Value }

func (*IntValue) schemaValueType() Type { return &Int{} }
