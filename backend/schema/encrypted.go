package schema

import (
	"fmt"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Encrypted struct {
	Pos  Position `parser:"" protobuf:"1,optional"`
	Type Type     `parser:"@@" protobuf:"2,optional"`
}

var _ Type = (*Encrypted)(nil)
var _ Symbol = (*Encrypted)(nil)

func (e *Encrypted) schemaSymbol()         {}
func (e *Encrypted) Equal(other Type) bool { return false }
func (e *Encrypted) Position() Position    { return e.Pos }
func (e *Encrypted) String() string        { return fmt.Sprintf("Encrypted<%v>", e.Type) }
func (e *Encrypted) ToProto() protoreflect.ProtoMessage {
	return &schemapb.Encrypted{Pos: posToProto(e.Pos), Type: TypeToProto(e.Type)}
}
func (e *Encrypted) schemaChildren() []Node { return []Node{e.Type} }
func (e *Encrypted) schemaType()            {}
