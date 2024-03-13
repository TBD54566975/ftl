package schema

import (
	"fmt"
	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

// RefProto is a constraint on the type of proto that can be used in a Ref.
type RefProto interface {
	schemapb.VerbRef | schemapb.DataRef | schemapb.SinkRef | schemapb.SourceRef | schemapb.EnumRef |
		schemapb.SecretRef | schemapb.ConfigRef
}

// Ref is an untyped reference to a symbol.
type Ref struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Module string `parser:"(@Ident '.')?" protobuf:"3"`
	Name   string `parser:"@Ident" protobuf:"2"`
	// Only used for data references.
	TypeParameters []Type `parser:"[ '<' @@ (',' @@)* '>' ]" protobuf:"4"`
}

// RefKey is a map key for a reference.
type RefKey struct {
	Module string
	Name   string
}

func (r Ref) ToRefKey() RefKey {
	return RefKey{Module: r.Module, Name: r.Name}
}
func (r *Ref) ToProto() proto.Message {
	panic("abstract ref must be translated to typed ref before proto conversion")
}

func (r *Ref) schemaChildren() []Node {
	out := make([]Node, 0, len(r.TypeParameters))
	for _, t := range r.TypeParameters {
		out = append(out, t)
	}
	return out
}

func (r *Ref) schemaType() {}

var _ Type = (*Ref)(nil)

func (r *Ref) Position() Position { return r.Pos }
func (r *Ref) String() string {
	out := makeRef(r.Module, r.Name)
	if len(r.TypeParameters) > 0 {
		out += "<"
		for i, t := range r.TypeParameters {
			if i != 0 {
				out += ", "
			}
			out += t.String()
		}
		out += ">"
	}
	return out
}

func RefFromProto(s *schemapb.Ref) *Ref {
	return &Ref{
		Pos:    posFromProto(s.Pos),
		Name:   s.Name,
		Module: s.Module,
	}
}

func ParseRef(ref string) (*Ref, error) {
	return refParser.ParseString("", ref)
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
