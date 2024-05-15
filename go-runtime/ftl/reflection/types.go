package reflection

import (
	"fmt"
	"strings"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

// Ref is an untyped reference to a symbol.
type Ref struct {
	Module string `json:"module"`
	Name   string `json:"name"`
}

func ParseRef(ref string) (Ref, error) {
	var out Ref
	if err := out.UnmarshalText([]byte(ref)); err != nil {
		return out, err
	}
	return out, nil
}

func (v *Ref) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), ".")
	if len(parts) != 2 {
		return fmt.Errorf("invalid reference %q", string(text))
	}
	v.Module = parts[0]
	v.Name = parts[1]
	return nil
}

func (v Ref) String() string { return v.Module + "." + v.Name }
func (v Ref) ToProto() *schemapb.Ref {
	return &schemapb.Ref{Module: v.Module, Name: v.Name}
}

func RefFromProto(p *schemapb.Ref) Ref { return Ref{Module: p.Module, Name: p.Name} }
