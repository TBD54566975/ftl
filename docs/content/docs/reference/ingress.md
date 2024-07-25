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
type GetRequest struct {
	UserID string `json:"userId"`
	PostID string `json:"postId"`
}

type GetResponse struct {
	Message string `json:"msg"`
}

//ftl:ingress GET /http/users/{userId}/posts
func Get(ctx context.Context, req builtin.HttpRequest[GetRequest]) (builtin.HttpResponse[GetResponse, ErrorResponse], error) {
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

Given the following request verb:

```go
type GetRequest struct {
	UserID string             `json:"userId"`
	Tag    ftl.Option[string] `json:"tag"`
	PostID string             `json:"postId"`
}

type GetResponse struct {
	Message string `json:"msg"`
}

//ftl:ingress http GET /users/{userId}/posts/{postId}
func Get(ctx context.Context, req builtin.HttpRequest[GetRequest]) (builtin.HttpResponse[GetResponse, string], error) {
	return builtin.HttpResponse[GetResponse, string]{
		Headers: map[string][]string{"Get": {"Header from FTL"}},
		Body: ftl.Some(GetResponse{
			Message: fmt.Sprintf("UserID: %s, PostID: %s, Tag: %s", req.Body.UserID, req.Body.PostID, req.Body.Tag.Default("none")),
		}),
	}, nil
}
```

`path`, `query`, and `body` parameters are automatically mapped to the `req` structure.

For example, this curl request will map `userId` to `req.Body.UserID` and `postId` to `req.Body.PostID`, and `tag` to `req.Body.Tag`:

```sh
curl -i http://localhost:8891/users/123/posts/456?tag=ftl
```

The response here will be:

```json
{
  "msg": "UserID: 123, PostID: 456, Tag: ftl"
}
```

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

//ftl:ingress http GET /typeenum
func TypeEnum(ctx context.Context, req builtin.HttpRequest[SumType]) (builtin.HttpResponse[SumType, string], error) {
	return builtin.HttpResponse[SumType, string]{Body: ftl.Some(req.Body)}, nil
}
```

The following curl request will map the `SumType` name and value to the `req.Body`:

```sh
curl -X GET "http://localhost:8891/typeenum" \
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
