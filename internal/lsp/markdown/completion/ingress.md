Declare an ingress function.

Verbs annotated with `ftl:ingress` will be exposed via HTTP (http is the default ingress type). These endpoints will then be available on one of our default ingress ports (local development defaults to http://localhost:8891).

```go
type GetPathParams struct {
	UserID string `json:"userId"`
}

type GetQueryParams struct {
	PostID string `json:"postId"`
}

type GetResponse struct {
	Message string `json:"msg"`
}

//ftl:ingress GET /http/users/{userId}/posts
func Get(ctx context.Context, req builtin.HttpRequest[ftl.Unit, GetPathParams, GetQueryParams]) (builtin.HttpResponse[GetResponse, string], error) {
  return builtin.HttpResponse[GetResponse, string]{
    Status:  200,
    Body:    ftl.Some(GetResponse{}),
  }, nil
}
```

See https://tbd54566975.github.io/ftl/docs/reference/ingress/
---

type ${1:Func}Request struct {
}

type ${1:Func}Response struct {
}

//ftl:ingress ${2:GET} ${3:/url/path}
func ${1:Func}(ctx context.Context, req builtin.HttpRequest[ftl.Unit, flt.Unit, ${1:Func}Request]) (builtin.HttpResponse[${1:Func}Response, string], error) {
	${4:// TODO: Implement}
	return builtin.HttpResponse[${1:Func}Response, string]{
		Status: 200,
		Body: ftl.Some(${1:Func}Response{}),
	}, nil
}
