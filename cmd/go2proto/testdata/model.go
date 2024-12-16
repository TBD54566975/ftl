package testdata

import (
	"net/url"
	"time"

	"github.com/block/ftl/internal/model"
)

type Root struct {
	Int            int                 `protobuf:"1"`
	String         string              `protobuf:"2"`
	MessagePtr     *Message            `protobuf:"4"`
	Enum           Enum                `protobuf:"5"`
	SumType        SumType             `protobuf:"6"`
	OptionalInt    int                 `protobuf:"7,optional"`
	OptionalIntPtr *int                `protobuf:"8,optional"`
	OptionalMsg    *Message            `protobuf:"9,optional"`
	RepeatedInt    []int               `protobuf:"10"`
	RepeatedMsg    []*Message          `protobuf:"11"`
	URL            *url.URL            `protobuf:"12"`
	Key            model.DeploymentKey `protobuf:"13"`
}

type Message struct {
	Time     time.Time     `protobuf:"1"`
	Duration time.Duration `protobuf:"2"`
	Nested   Nested        `protobuf:"3"`
}

type Nested struct {
	Nested string `protobuf:"1"`
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

type SubSumType interface {
	SumType

	subSumType()
}

//protobuf:1
//protobuf:4 SumType
type SubSumTypeA struct {
	A string `protobuf:"1"`
}

func (SubSumTypeA) subSumType() {}
func (SubSumTypeA) sumType()    {}

//protobuf:2
//protobuf:5 SumType
type SubSumTypeB struct {
	A string `protobuf:"1"`
}

func (SubSumTypeB) subSumType() {}
func (SubSumTypeB) sumType()    {}
