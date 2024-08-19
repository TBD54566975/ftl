+++
title = "HTTP Ingress"
description = "Handling incoming HTTP requests"
date = 2021-05-01T08:20:00+00:00
updated = 2021-05-01T08:20:00+00:00
draft = false
weight = 50
sort_by = "weight"
template = "docs/page.html"

[extra]
toc = true
top = false
+++

Verbs annotated with `ftl:ingress` will be exposed via HTTP (`http` is the default ingress type). These endpoints will then be available on one of our default `ingress` ports (local development defaults to `http://localhost:8891`).

The following will be available at `http://localhost:8891/http/users/123/posts?postId=456`.

```go
type GetRequestPathParams struct {
	UserID string `json:"userId"`
}

type GetRequestQueryParams struct {
    PostID string `json:"postId"`
}

type GetResponse struct {
	Message string `json:"msg"`
}

//ftl:ingress GET /http/users/{userId}/posts
func Get(ctx context.Context, req builtin.HttpRequest[ftl.Unit, GetRequestPathParams, GetRequestQueryParams]) (builtin.HttpResponse[GetResponse, ErrorResponse], error) {
  // ...
}
```

Because the example above only has a single path parameter it can be simplified by just using a scalar such as `string` or `int64` as the path parameter type:

```go

//ftl:ingress GET /http/users/{userId}/posts
func Get(ctx context.Context, req builtin.HttpRequest[ftl.Unit, int64, GetRequestQueryParams]) (builtin.HttpResponse[GetResponse, ErrorResponse], error) {
  // ...
}
```

> **NOTE!**
> The `req` and `resp` types of HTTP `ingress` [verbs](../verbs) must be `builtin.HttpRequest` and `builtin.HttpResponse` respectively. These types provide the necessary fields for HTTP `ingress` (`headers`, `statusCode`, etc.)
>
> You will need to import `ftl/builtin`.

Key points:

- `ingress` verbs will be automatically exported by default.

## Field mapping

The `HttpRequest` request object takes 3 type parameters, the body, the path parameters and the query parameters.

Given the following request verb:

```go

type PostBody struct{
    Title string               `json:"title"`
	Content string             `json:"content"`
    Tag ftl.Option[string]     `json:"tag"`
}
type PostPathParams struct {
	UserID string             `json:"userId"`
	PostID string             `json:"postId"`
}

type PostQueryParams struct {
	Publish boolean `json:"publish"`
}

//ftl:ingress http PUT /users/{userId}/posts/{postId}
func Get(ctx context.Context, req builtin.HttpRequest[PostBody, PostPathParams, PostQueryParams]) (builtin.HttpResponse[GetResponse, string], error) {
	return builtin.HttpResponse[GetResponse, string]{
		Headers: map[string][]string{"Get": {"Header from FTL"}},
		Body: ftl.Some(GetResponse{
			Message: fmt.Sprintf("UserID: %s, PostID: %s, Tag: %s", req.pathParameters.UserID, req.pathParameters.PostID, req.Body.Tag.Default("none")),
		}),
	}, nil
}
```

The rules for how each element is mapped are slightly different, as they have a different structure:

- The body is mapped directly to the body of the request, generally as a JSON object. Scalars are also supported, as well as []byte to get the raw body. If they type is `any` then it will be assumed to be JSON and mapped to the appropriate types based on the JSON structure.
- The path parameters can be mapped directly to an object with field names corresponding to the name of the path parameter. If there is only a single path parameter it can be injected directly as a scalar. They can also be injected as a `map[string]string`.
- The path parameters can also be mapped directly to an object with field names corresponding to the name of the path parameter. They can also be injected directly as a `map[string]string`, or `map[string][]string` for multiple values.

#### Optional fields

Optional fields are represented by the `ftl.Option` type. The `Option` type is a wrapper around the actual type and can be `Some` or `None`. In the example above, the `Tag` field is optional.

```sh
curl -i http://localhost:8891/users/123/posts/456
```

Because the `tag` query parameter is not provided, the response will be:

```json
{
  "msg": "UserID: 123, PostID: 456, Tag: none"
}
```

#### Casing

Field names use lowerCamelCase by default. You can override this by using the `json` tag.

## SumTypes

Given the following request verb:

```go
//ftl:enum export
type SumType interface {
	tag()
}

type A string

func (A) tag() {}

type B []string

func (B) tag() {}

//ftl:ingress http POST /typeenum
func TypeEnum(ctx context.Context, req builtin.HttpRequest[SumType, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[SumType, string], error) {
	return builtin.HttpResponse[SumType, string]{Body: ftl.Some(req.Body)}, nil
}
```

The following curl request will map the `SumType` name and value to the `req.Body`:

```sh
curl -X POST "http://localhost:8891/typeenum" \
     -H "Content-Type: application/json" \
     --data '{"name": "A", "value": "sample"}'
```

The response will be:

```json
{
  "name": "A",
  "value": "sample"
}
```

## Encoding query params as JSON

Complex query params can also be encoded as JSON using the `@json` query parameter. For example:

> `{"tag":"ftl"}` url-encoded is `%7B%22tag%22%3A%22ftl%22%7D`

```bash
curl -i http://localhost:8891/users/123/posts/456?@json=%7B%22tag%22%3A%22ftl%22%7D
```
