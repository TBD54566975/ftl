Snippet for defining a verb function.

```go
//ftl:verb
func Name(ctx context.Context, req Request) (Response, error) {}
```
---
type ${1:Request} struct {}
type ${2:Response} struct {}

//ftl:verb
func ${3:Name}(ctx context.Context, req ${1:Request}) (${2:Response}, error) {
return ${2:Response}{}, nil
}