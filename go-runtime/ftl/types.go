package ftl

import (
	"context"
)

// Handle represents a resource that can be retrieved such as a database connection, secret, etc.
type Handle[T any] interface {
	Get(ctx context.Context) T
}

// Unit is a type that has no value.
//
// It can be used as a parameter or return value to indicate that a function
// does not accept or return any value.
type Unit struct{}

// A Verb is a function that accepts input and returns output.
type Verb[Req, Resp any] func(context.Context, Req) (Resp, error)

// A Sink is a function that accepts input but returns nothing.
type Sink[Req any] func(context.Context, Req) error

// A Source is a function that does not accept input but returns output.
type Source[Resp any] func(context.Context) (Resp, error)

// An Empty is a function that does not accept input or return output.
type Empty func(context.Context) error
