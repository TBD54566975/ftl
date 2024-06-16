Snippet for declaring a verb function.

```go
//ftl:verb
func Name(ctx context.Context, req Request) (Response, error) {}
```

See https://tbd54566975.github.io/ftl/docs/reference/verbs/
---
type ${1:Request} struct {}
type ${2:Response} struct {}

//ftl:verb
func ${3:Name}(ctx context.Context, req ${1:Request}) (${2:Response}, error) {
	${4:// TODO: Implement}
	return ${2:Response}{}, nil
}