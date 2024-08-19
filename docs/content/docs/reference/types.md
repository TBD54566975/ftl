+++
title = "Types"
description = "Declaring and using Types"
date = 2021-05-01T08:20:00+00:00
updated = 2021-05-01T08:20:00+00:00
draft = false
weight = 30
sort_by = "weight"
template = "docs/page.html"

[extra]
toc = true
top = false
+++

FTL supports the following types: `Int` (64-bit), `Float` (64-bit), `String`, `Bytes` (a byte array), `Bool`, `Time`, `Any` (a dynamic type), `Unit` (similar to "void"), arrays, maps, data structures, and constant enumerations. Each FTL type is mapped to a corresponding language-specific type. For example in Go `Float` is represented as `float64`, `Time` is represented by `time.Time`, and so on. [^1]

Any Go type supported by FTL and referenced by an FTL declaration will be automatically exposed to an FTL type.

For example, the following verb declaration will result in `Request` and `Response` being automatically translated to FTL types.

```go
type Request struct {}
type Response struct {}

//ftl:verb
func Hello(ctx context.Context, in Request) (Response, error) {
  // ...
}
```

## Type enums (sum types)

[Sum types](https://en.wikipedia.org/wiki/Tagged_union) are supported by FTL's type system, but aren't directly supported by Go. However they can be approximated with the use of [sealed interfaces](https://blog.chewxy.com/2018/03/18/golang-interfaces/). To declare a sum type in FTL use the comment directive `//ftl:enum`:

```go
//ftl:enum
type Animal interface { animal() }

type Cat struct {}
func (Cat) animal() {}

type Dog struct {}
func (Dog) animal() {}
```

## Value enums

A value enum is an enumerated set of string or integer values.

```go
//ftl:enum
type Colour string

const (
  Red   Colour = "red"
  Green Colour = "green"
  Blue  Colour = "blue"
)
```

## Type aliases

A type alias is an alternate name for an existing type. It can be declared like so:

```go
//ftl:typealias
type Alias Target
```
or
```go
//ftl:typealias
type Alias = Target
```

eg.

```go
//ftl:typealias
type UserID string

//ftl:typealias
type UserToken = string
```

---

## Optional types

FTL supports optional types, which are types that can be `None` or `Some` and can be declared via `ftl.Option[T]`. These types are provided by the `ftl` runtimes. For example, the following FTL type declaration in go, will provide an optional string type "Name":

```go
type EchoResponse struct {
	Name ftl.Option[string] `json:"name"`
}
```

The value of this type can be set to `Some` or `None`:

```go
resp := EchoResponse{
  Name: ftl.Some("John"),
}

resp := EchoResponse{
  Name: ftl.None(),
}
```

The value of the optional type can be accessed using `Get`, `MustGet`, or `Default` methods:

```go
// Get returns the value and a boolean indicating if the Option contains a value.
if value, ok := resp.Name.Get(); ok {
  resp.Name = ftl.Some(value)
}

// MustGet returns the value or panics if the Option is None.
value := resp.Name.MustGet()

// Default returns the value or a default value if the Option is None.
value := resp.Name.Default("default")
``` 

## Unit "void" type

The `Unit` type is similar to the `void` type in other languages. It is used to indicate that a function does not return a value. For example:

```go
//ftl:ingress GET /unit
func Unit(ctx context.Context, req builtin.HttpRequest[ftl.Unit, ftl.Unit, TimeRequest]) (builtin.HttpResponse[ftl.Unit, string], error) {
	return builtin.HttpResponse[ftl.Unit, string]{Body: ftl.Some(ftl.Unit{})}, nil
}
```

This request will return an empty body with a status code of 200:

```sh
curl http://localhost:8891/unit -i
```

```http
HTTP/1.1 200 OK
Date: Mon, 12 Aug 2024 17:58:22 GMT
Content-Length: 0
```

## Builtin types

FTL provides a set of builtin types that are automatically available in all FTL runtimes. These types are:

- `builtin.HttpRequest[Body, PathParams, QueryParams]` - Represents an HTTP request with a body of type `Body`, path parameter type of `PathParams` and a query parameter type of `QueryParams`.
- `builtin.HttpResponse[Body, Error]` - Represents an HTTP response with a body of type `Body` and an error of type `Error`.
- `builtin.Empty` - Represents an empty type. This equates to an empty structure `{}`.
- `builtin.CatchRequest` - Represents a request structure for catch verbs.

