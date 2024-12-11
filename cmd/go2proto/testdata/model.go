package testdata

import (
	"time"
)

type Root struct {
	Int            int        `protobuf:"1"`
	String         string     `protobuf:"2"`
	MessagePtr     *Message   `protobuf:"4"`
	Enum           Enum       `protobuf:"5"`
	SumType        SumType    `protobuf:"6"`
	OptionalInt    int        `protobuf:"7,optional"`
	OptionalIntPtr *int       `protobuf:"8,optional"`
	OptionalMsg    *Message   `protobuf:"9,optional"`
	RepeatedInt    []int      `protobuf:"10"`
	RepeatedMsg    []*Message `protobuf:"11"`
}

type Message struct {
	Time     time.Time     `protobuf:"1"`
	Duration time.Duration `protobuf:"2"`
}

type Enum int

const (
	EnumA Enum = iota
	EnumB
)

type SumType interface {
	sumType()
}

//protobuf:1
type SumTypeA struct {
	A string `protobuf:"1"`
}

func (SumTypeA) sumType() {}

//protobuf:2
type SumTypeB struct {
	B int `protobuf:"1"`
}

func (SumTypeB) sumType() {}

//protobuf:3
type SumTypeC struct {
	C float64 `protobuf:"1"`
}

func (SumTypeC) sumType() {}
