package schema

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

// RefProto is a constraint on the type of proto that can be used in a Ref.
type RefProto interface {
	schemapb.VerbRef | schemapb.DataRef | schemapb.SinkRef | schemapb.SourceRef | schemapb.EnumRef
}

// Ref is an untyped reference to a symbol.
type Ref struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Module string `parser:"(@Ident '.')?" protobuf:"3"`
	Name   string `parser:"@Ident" protobuf:"2"`
}

func (b *Ref) Position() Position { return b.Pos }
func (b *Ref) String() string     { return makeRef(b.Module, b.Name) }

// AbstractRef is an abstract reference to a function or data type.
type AbstractRef[Proto RefProto] Ref

func ParseRef[Proto RefProto](ref string) (*AbstractRef[Proto], error) {
	parts := strings.Split(ref, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid reference %q", ref)
	}
	return &AbstractRef[Proto]{Module: parts[0], Name: parts[1]}, nil
}

// Untyped converts a typed reference to an untyped reference.
func (a *AbstractRef[Proto]) Untyped() Ref       { return Ref(*a) }
func (a *AbstractRef[Proto]) Position() Position { return a.Pos }
func (a *AbstractRef[Proto]) ToProto() proto.Message {
	switch any((*Proto)(nil)).(type) {
	case *schemapb.VerbRef:
		return any(&schemapb.VerbRef{Pos: posToProto(a.Pos), Module: a.Module, Name: a.Name}).(proto.Message) //nolint:forcetypeassert

	case *schemapb.DataRef:
		return any(&schemapb.DataRef{Pos: posToProto(a.Pos), Module: a.Module, Name: a.Name}).(proto.Message) //nolint:forcetypeassert

	case *schemapb.SinkRef:
		return any(&schemapb.SinkRef{Pos: posToProto(a.Pos), Module: a.Module, Name: a.Name}).(proto.Message) //nolint:forcetypeassert

	case *schemapb.SourceRef:
		return any(&schemapb.SourceRef{Pos: posToProto(a.Pos), Module: a.Module, Name: a.Name}).(proto.Message) //nolint:forcetypeassert

	case *schemapb.EnumRef:
		return any(&schemapb.EnumRef{Pos: posToProto(a.Pos), Module: a.Module, Name: a.Name}).(proto.Message) //nolint:forcetypeassert

	default:
		panic(fmt.Sprintf("unsupported ref proto type %T", (*Proto)(nil)))
	}
}

func (a *AbstractRef[Proto]) String() string       { return makeRef(a.Module, a.Name) }
func (*AbstractRef[Proto]) schemaChildren() []Node { return nil }
func (*AbstractRef[Proto]) schemaType()            {}
