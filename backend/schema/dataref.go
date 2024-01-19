package schema

import (
	"google.golang.org/protobuf/reflect/protoreflect"

	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

// DataRef is a reference to a data structure.
type DataRef struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Module         string `parser:"(@Ident '.')?" protobuf:"3"`
	Name           string `parser:"@Ident" protobuf:"2"`
	TypeParameters []Type `parser:"[ '<' @@ (',' @@)* '>' ]" protobuf:"4"`
}

var _ Type = (*DataRef)(nil)

func (d *DataRef) Position() Position { return d.Pos }

// Untyped converts a typed reference to an untyped reference.
func (d *DataRef) Untyped() Ref {
	return Ref{Pos: d.Pos, Module: d.Module, Name: d.Name}
}

func (d *DataRef) String() string {
	out := d.Name
	if d.Module != "" {
		out = d.Module + "." + out
	}
	if len(d.TypeParameters) > 0 {
		out += "<"
		for i, t := range d.TypeParameters {
			if i != 0 {
				out += ", "
			}
			out += t.String()
		}
		out += ">"
	}
	return out
}

func (d *DataRef) ToProto() protoreflect.ProtoMessage {
	return &schemapb.DataRef{Pos: posToProto(d.Pos), Module: d.Module, Name: d.Name}
}

func (*DataRef) schemaChildren() []Node { return nil }
func (*DataRef) schemaType()            {}

func ParseDataRef(ref string) (*DataRef, error) {
	return dataRefParser.ParseString("", ref)
}

func DataRefFromProto(s *schemapb.DataRef) *DataRef {
	return &DataRef{
		Pos:    posFromProto(s.Pos),
		Name:   s.Name,
		Module: s.Module,
	}
}

func dataRefToSchema(s *schemapb.DataRef) *DataRef {
	return &DataRef{
		Pos:    posFromProto(s.Pos),
		Name:   s.Name,
		Module: s.Module,
	}
}
