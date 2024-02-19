package ftl

import (
	"context"
	"fmt"
	"strings"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
)

// Unit is a type that has no value.
//
// It can be used as a parameter or return value to indicate that a function
// does not accept or return any value.
type Unit struct{}

// Ref is an untyped reference to a symbol.
type Ref struct {
	Module string `json:"module"`
	Name   string `json:"name"`
}

// AbstractRef is an abstract reference to a symbol.
type AbstractRef[Proto schema.RefProto] Ref

func ParseRef[Proto schema.RefProto](ref string) (AbstractRef[Proto], error) {
	var out AbstractRef[Proto]
	if err := out.UnmarshalText([]byte(ref)); err != nil {
		return out, err
	}
	return out, nil
}

func (v *AbstractRef[Proto]) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid reference %q", string(text))
	}
	v.Module = parts[0]
	v.Name = parts[1]
	return nil
}

func (v *AbstractRef[Proto]) String() string { return v.Module + "." + v.Name }
func (v *AbstractRef[Proto]) ToProto() *Proto {
	switch any((*Proto)(nil)).(type) {
	case *schemapb.VerbRef:
		return any(&schemapb.VerbRef{Module: v.Module, Name: v.Name}).(*Proto) //nolint:forcetypeassert

	case *schemapb.DataRef:
		return any(&schemapb.DataRef{Module: v.Module, Name: v.Name}).(*Proto) //nolint:forcetypeassert

	case *schemapb.SinkRef:
		return any(&schemapb.SinkRef{Module: v.Module, Name: v.Name}).(*Proto) //nolint:forcetypeassert

	case *schemapb.SourceRef:
		return any(&schemapb.SourceRef{Module: v.Module, Name: v.Name}).(*Proto) //nolint:forcetypeassert

	default:
		panic(fmt.Sprintf("unsupported proto type %T", (*Proto)(nil)))
	}
}

// A Verb is a function that accepts input and returns output.
type Verb[Req, Resp any] func(context.Context, Req) (Resp, error)

// VerbRef is a reference to a verb (a function in the form F(I)O).
type VerbRef = AbstractRef[schemapb.VerbRef]

func ParseVerbRef(ref string) (VerbRef, error)     { return ParseRef[schemapb.VerbRef](ref) }
func VerbRefFromProto(p *schemapb.VerbRef) VerbRef { return VerbRef{Module: p.Module, Name: p.Name} }

// A Sink is a function that accepts input but returns nothing.
type Sink[Req any] func(context.Context, Req) error

type SinkRef = AbstractRef[schemapb.SinkRef]

func ParseSinkRef(ref string) (SinkRef, error)     { return ParseRef[schemapb.SinkRef](ref) }
func SinkRefFromProto(p *schemapb.SinkRef) SinkRef { return SinkRef{Module: p.Module, Name: p.Name} }

// A Source is a function that does not accept input but returns output.
type Source[Req any] func(context.Context, Req) error

type SourceRef = AbstractRef[schemapb.SourceRef]

func ParseSourceRef(ref string) (SourceRef, error) { return ParseRef[schemapb.SourceRef](ref) }
func SourceRefFromProto(p *schemapb.SourceRef) SourceRef {
	return SourceRef{Module: p.Module, Name: p.Name}
}

// DataRef is a reference to a Data type.
type DataRef = AbstractRef[schemapb.DataRef]

func ParseDataRef(ref string) (DataRef, error)     { return ParseRef[schemapb.DataRef](ref) }
func DataRefFromProto(p *schemapb.DataRef) DataRef { return DataRef{Module: p.Module, Name: p.Name} }
