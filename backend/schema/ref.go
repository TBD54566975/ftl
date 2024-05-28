package schema

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"google.golang.org/protobuf/proto"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/internal/slices"
)

// RefKey is a map key for a reference.
type RefKey struct {
	Module string `parser:"(@Ident '.')?"`
	Name   string `parser:"@Ident"`
}

func (r RefKey) ToRef() *Ref                  { return &Ref{Module: r.Module, Name: r.Name} }
func (r RefKey) String() string               { return makeRef(r.Module, r.Name) }
func (r RefKey) ToProto() *schemapb.Ref       { return &schemapb.Ref{Module: r.Module, Name: r.Name} }
func (r RefKey) Value() (driver.Value, error) { return r.String(), nil }
func (r *RefKey) Scan(src any) error {
	p, err := ParseRef(src.(string))
	if err != nil {
		return fmt.Errorf("%v: %w", src, err)
	}
	*r = p.ToRefKey()
	return nil
}

// Ref is an untyped reference to a symbol.
type Ref struct {
	Pos Position `parser:"" protobuf:"1,optional"`

	Module string `parser:"(@Ident '.')?" protobuf:"3"`
	Name   string `parser:"@Ident" protobuf:"2"`
	// Only used for data references.
	TypeParameters []Type `parser:"[ '<' @@ (',' @@)* '>' ]" protobuf:"4"`
}

var _ sql.Scanner = (*Ref)(nil)
var _ driver.Valuer = (*Ref)(nil)

func (r Ref) Value() (driver.Value, error) { return r.String(), nil }

func (r *Ref) Scan(src any) error {
	p, err := ParseRef(src.(string))
	if err != nil {
		return err
	}
	*r = *p
	return nil
}

func (r Ref) ToRefKey() RefKey {
	return RefKey{Module: r.Module, Name: r.Name}
}

func (r *Ref) ToProto() proto.Message {
	return &schemapb.Ref{
		Pos:            posToProto(r.Pos),
		Module:         r.Module,
		Name:           r.Name,
		TypeParameters: slices.Map(r.TypeParameters, TypeToProto),
	}
}

func (r *Ref) Equal(other Type) bool {
	or, ok := other.(*Ref)
	if !ok {
		return false
	}
	if r.Module != or.Module || r.Name != or.Name || len(r.TypeParameters) != len(or.TypeParameters) {
		return false
	}
	for i, t := range r.TypeParameters {
		if !t.Equal(or.TypeParameters[i]) {
			return false
		}
	}
	return true
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
		Pos:            posFromProto(s.Pos),
		Name:           s.Name,
		Module:         s.Module,
		TypeParameters: slices.Map(s.TypeParameters, TypeFromProto),
	}
}

func ParseRef(ref string) (*Ref, error) {
	out, err := refParser.ParseString("", ref)
	if err != nil {
		return nil, err
	}
	out.Pos = Position{}
	return out, nil
}

func refListToSchema(s []*schemapb.Ref) []*Ref {
	var out []*Ref
	for _, n := range s {
		out = append(out, RefFromProto(n))
	}
	return out
}
