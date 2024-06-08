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

eg.

```go
//ftl:typealias
type UserID string
```

[^1]: Note that until [type widening](https://github.com/TBD54566975/ftl/issues/1296) is implemented, external types are not supported.
