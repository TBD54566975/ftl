package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

// RefProto is a constraint on the type of proto that can be used in a Ref.
type RefProto interface {
	schemapb.VerbRef | schemapb.DataRef | schemapb.SinkRef | schemapb.SourceRef
}

// Ref is an untyped reference to a symbol.
type Ref struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Module string `parser:"(@Ident '.')?" protobuf:"3"`
	Name   string `parser:"@Ident" protobuf:"2"`
}

func (b *Ref) String() string { return makeRef(b.Module, b.Name) }

// AbstractRef is an abstract reference to a function or data type.
type AbstractRef[Proto RefProto] Ref

func ParseRef[Proto RefProto](ref string) (*AbstractRef[Proto], error) {
	parts := strings.Split(ref, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid reference %q", ref)
	}
	return &AbstractRef[Proto]{Module: parts[0], Name: parts[1]}, nil
}

func (v *AbstractRef[Proto]) ToProto() proto.Message {
	switch any((*Proto)(nil)).(type) {
	case *schemapb.VerbRef:
		return any(&schemapb.VerbRef{Pos: posToProto(v.Pos), Module: v.Module, Name: v.Name}).(proto.Message) //nolint:forcetypeassert

	case *schemapb.DataRef:
		return any(&schemapb.DataRef{Pos: posToProto(v.Pos), Module: v.Module, Name: v.Name}).(proto.Message) //nolint:forcetypeassert

	case *schemapb.SinkRef:
		return any(&schemapb.SinkRef{Pos: posToProto(v.Pos), Module: v.Module, Name: v.Name}).(proto.Message) //nolint:forcetypeassert

	case *schemapb.SourceRef:
		return any(&schemapb.SourceRef{Pos: posToProto(v.Pos), Module: v.Module, Name: v.Name}).(proto.Message) //nolint:forcetypeassert

	default:
		panic(fmt.Sprintf("unsupported ref proto type %T", (*Proto)(nil)))
	}
}

func (v AbstractRef[Proto]) String() string        { return makeRef(v.Module, v.Name) }
func (*AbstractRef[Proto]) schemaChildren() []Node { return nil }
func (*AbstractRef[Proto]) schemaType()            {}
