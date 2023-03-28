package sdkgo

import (
	"strings"

	"github.com/alecthomas/errors"
	"google.golang.org/protobuf/proto"

	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

// Ref is a reference to a Verb or Data.
type Ref[Proto proto.Message] struct {
	Module string `json:"module,omitempty"`
	Name   string `json:"name"`
}

type VerbRef = Ref[*pschema.VerbRef]
type DataRef = Ref[*pschema.DataRef]

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
