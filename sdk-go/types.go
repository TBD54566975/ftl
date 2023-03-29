package sdkgo

import (
	"context"
	"strings"

	"github.com/alecthomas/errors"
	"google.golang.org/protobuf/proto"

	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

// A Verb is a function that can be called with an input and an output.
type Verb[Req, Resp any] func(context.Context, Req) (Resp, error)

// A Sink is a function that can be called with an input and no output.
type Sink[Req any] func(context.Context, Req) error

// Ref is a reference to a Verb or Data.
type Ref[Proto proto.Message] struct {
	Module string `json:"module,omitempty"`
	Name   string `json:"name"`
}

// VerbRef is a reference to a Verb.
type VerbRef = Ref[*pschema.VerbRef]

// DataRef is a reference to a Data type.
type DataRef = Ref[*pschema.DataRef]

// ParseRef parses a reference from a string.
func ParseRef[Proto proto.Message](ref string) (Ref[Proto], error) {
	var out Ref[Proto]
	return out, out.UnmarshalText([]byte(ref))
}

func ParseVerbRef(ref string) (VerbRef, error)    { return ParseRef[*pschema.VerbRef](ref) }
func VerbRefFromProto(p *pschema.VerbRef) VerbRef { return VerbRef{Module: p.Module, Name: p.Name} }
func ParseDataRef(ref string) (DataRef, error)    { return ParseRef[*pschema.DataRef](ref) }
func DataRefFromProto(p *pschema.DataRef) DataRef { return DataRef{Module: p.Module, Name: p.Name} }

func (v *Ref[Proto]) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), ".")
	if len(parts) != 2 {
		return errors.Errorf("invalid reference %q", string(text))
	}
	v.Module = parts[0]
	v.Name = parts[1]
	return nil
}

func (v Ref[Proto]) String() string {
	return v.Module + "." + v.Name
}

func (v Ref[Proto]) ToProto() Proto {
	var p Proto
	switch p := any(p).(type) {
	case *pschema.VerbRef:
		p.Module = v.Module
		p.Name = v.Name
	case *pschema.DataRef:
		p.Module = v.Module
		p.Name = v.Name
	default:
		panic("???")
	}
	return p
}
