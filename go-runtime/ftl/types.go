package ftl

import (
	"context"
	"fmt"
	"strings"

	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
)

// Handle represents a resource that can be retrieved such as a database connection, secret, etc.
type Handle[T any] interface {
	Get(ctx context.Context) T
}

// HashableHandle is a Handle that can be hashed to determine when it's value has changed.
type HashableHandle[T any] interface {
	Handle[T]
	Hash(ctx context.Context) []byte
}

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

// A Verb is a function that accepts input and returns output.
type Verb[Req, Resp any] func(context.Context, Req) (Resp, error)

// A Sink is a function that accepts input but returns nothing.
type Sink[Req any] func(context.Context, Req) error

// A Source is a function that does not accept input but returns output.
type Source[Resp any] func(context.Context) (Resp, error)

// An Empty is a function that does not accept input or return output.
type Empty func(context.Context) error
