package sdkgo

import (
	"context"
	"strings"

	"github.com/alecthomas/errors"

	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

// A Verb is a function that can be called with an input and an output.
type Verb[Req, Resp any] func(context.Context, Req) (Resp, error)

// A Sink is a function that can be called with an input and no output.
type Sink[Req any] func(context.Context, Req) error

func ParseVerbRef(ref string) (VerbRef, error) {
	m, n, err := parseRef(ref)
	if err != nil {
		return VerbRef{}, err
	}
	return VerbRef{Module: m, Name: n}, nil
}
func VerbRefFromProto(p *pschema.VerbRef) VerbRef { return VerbRef{Module: p.Module, Name: p.Name} }

// VerbRef is a reference to a Verb.
type VerbRef struct {
	Module string `json:"module,omitempty"`
	Name   string `json:"name"`
}

func (v *VerbRef) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), ".")
	if len(parts) != 2 {
		return errors.Errorf("invalid reference %q", string(text))
	}
	v.Module = parts[0]
	v.Name = parts[1]
	return nil
}

func (v VerbRef) String() string            { return v.Module + "." + v.Name }
func (v VerbRef) ToProto() *pschema.VerbRef { return &pschema.VerbRef{Module: v.Module, Name: v.Name} }

func ParseDataRef(ref string) (DataRef, error) {
	m, n, err := parseRef(ref)
	if err != nil {
		return DataRef{}, err
	}
	return DataRef{Module: m, Name: n}, nil
}
func DataRefFromProto(p *pschema.DataRef) DataRef { return DataRef{Module: p.Module, Name: p.Name} }

// DataRef is a reference to a Data type.
type DataRef struct {
	Module string `json:"module,omitempty"`
	Name   string `json:"name"`
}

func (v *DataRef) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), ".")
	if len(parts) != 2 {
		return errors.Errorf("invalid reference %q", string(text))
	}
	v.Module = parts[0]
	v.Name = parts[1]
	return nil
}

func (v DataRef) String() string            { return v.Module + "." + v.Name }
func (v DataRef) ToProto() *pschema.DataRef { return &pschema.DataRef{Module: v.Module, Name: v.Name} }

func parseRef(s string) (string, string, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return "", "", errors.Errorf("invalid reference %q", s)
	}
	return parts[0], parts[1], nil
}
