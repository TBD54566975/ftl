## Verb

Verbs are the function primitives of FTL. They take a single value and return a single value or an error.

`F(X) -> Y`

eg.

```go
type EchoRequest struct {}
type EchoResponse struct {}

//ftl:verb
func Echo(ctx context.Context, in EchoRequest) (EchoResponse, error) {
  // ...
}
```

[Reference](https://tbd54566975.github.io/ftl/docs/reference/verbs/)
