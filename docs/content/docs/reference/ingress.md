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

Key points to note

- `path`, `query`, and `body` parameters are automatically mapped to the `req` and `resp` structures. In the example above, `{userId}` is extracted from the path parameter and `postId` is extracted from the query parameter.
- `ingress` verbs will be automatically exported by default.
