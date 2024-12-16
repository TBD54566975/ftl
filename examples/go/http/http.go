package http

import (
	"context"
	"ftl/builtin"

	"github.com/block/ftl/go-runtime/ftl"
)

type ApiError struct {
	Message string `json:"message"`
}

type GetQueryParams struct {
	Age ftl.Option[int] `json:"age"`
}

type GetPathParams struct {
	Name string `json:"name"`
}

type GetResponse struct {
	Name string          `json:"name"`
	Age  ftl.Option[int] `json:"age"`
}

// Example usage of path and query params
// curl http://localhost:8891/get/wicket?age=10
//
//ftl:ingress http GET /get/{name}
func Get(ctx context.Context, req builtin.HttpRequest[ftl.Unit, GetPathParams, GetQueryParams]) (builtin.HttpResponse[GetResponse, ApiError], error) {
	return builtin.HttpResponse[GetResponse, ApiError]{
		Status: 200,
		Body: ftl.Some(GetResponse{
			Name: req.PathParameters.Name,
			Age:  req.Query.Age,
		}),
	}, nil
}

type PostRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type PostResponse struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// Example POST request with a JSON body
// curl -X POST http://localhost:8891/post -d '{"name": "wicket", "age": 10}'
//
//ftl:ingress http POST /post
func Post(ctx context.Context, req builtin.HttpRequest[PostRequest, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[PostResponse, ApiError], error) {
	return builtin.HttpResponse[PostResponse, ApiError]{
		Status: 200,
		Body: ftl.Some(PostResponse{
			Name: req.Body.Name,
			Age:  req.Body.Age,
		}),
	}, nil
}
