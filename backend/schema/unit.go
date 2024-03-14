package schema

import (
	"google.golang.org/protobuf/reflect/protoreflect"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

type Unit struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Unit bool `parser:"@'Unit'" protobuf:"-"`
}

var _ Type = (*Unit)(nil)
var _ Symbol = (*Unit)(nil)

func (u *Unit) Position() Position                 { return u.Pos }
func (u *Unit) schemaType()                        {}
func (u *Unit) schemaSymbol()                      {}
func (u *Unit) String() string                     { return "Unit" }
func (u *Unit) ToProto() protoreflect.ProtoMessage { return &schemapb.Unit{Pos: posToProto(u.Pos)} }
func (u *Unit) schemaChildren() []Node             { return nil }
func (u *Unit) GetName() string                    { return "" }
