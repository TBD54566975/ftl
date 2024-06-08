+++
title = "Verbs"
description = "Declaring and calling Verbs"
date = 2021-05-01T08:20:00+00:00
updated = 2021-05-01T08:20:00+00:00
draft = false
weight = 20
sort_by = "weight"
template = "docs/page.html"

[extra]
toc = true
top = false
+++

## Defining Verbs

To declare a Verb, write a normal Go function with the following signature,annotated with the Go [comment directive](https://tip.golang.org/doc/comment#syntax) `//ftl:verb`:

```go
//ftl:verb
func F(context.Context, In) (Out, error) { }
```

eg.

```go
type EchoRequest struct {}

type EchoResponse struct {}

//ftl:verb
func Echo(ctx context.Context, in EchoRequest) (EchoResponse, error) {
  // ...
}
```

By default verbs are only [visible](../visibility) to other verbs in the same module.

## Calling Verbs

To call a verb use `ftl.Call()`. eg.

```go
out, err := ftl.Call(ctx, echo.Echo, echo.EchoRequest{})
```
