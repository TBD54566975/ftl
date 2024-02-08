//ftl:module echo
package echo

import (
	"context"
	"fmt"

	"ftl/builtin"

	ftl "github.com/TBD54566975/ftl/go-runtime/sdk" // Import the FTL SDK.
)

type GetRequest struct {
	UserID string `alias:"userId"`
	PostID string `alias:"postId"`
}

type Nested struct {
	GoodStuff string `alias:"good_stuff"`
}

type GetResponse struct {
	Message string `alias:"msg"`
	Nested  Nested `alias:"nested"`
}

//ftl:verb
//ftl:ingress http GET /users/{userId}/posts/{postId}
func Get(ctx context.Context, req builtin.HttpRequest[GetRequest]) (builtin.HttpResponse[GetResponse], error) {
	return builtin.HttpResponse[GetResponse]{
		Headers: map[string][]string{"Get": {"Header from FTL"}},
		Body: GetResponse{
			Message: fmt.Sprintf("UserID: %s, PostID: %s", req.Body.UserID, req.Body.PostID),
			Nested: Nested{
				GoodStuff: "This is good stuff",
			},
		},
	}, nil
}

type PostRequest struct {
	UserID int `alias:"user_id"`
	PostID int
}

type PostResponse struct {
	Success bool `alias:"success"`
}

//ftl:verb
//ftl:ingress http POST /users
func Post(ctx context.Context, req builtin.HttpRequest[PostRequest]) (builtin.HttpResponse[PostResponse], error) {
	return builtin.HttpResponse[PostResponse]{
		Status:  201,
		Headers: map[string][]string{"Post": {"Header from FTL"}},
		Body:    PostResponse{Success: true},
	}, nil
}

type PutRequest struct {
	UserID string `alias:"userId"`
	PostID string `alias:"postId"`
}

type PutResponse struct{}

//ftl:verb
//ftl:ingress http PUT /users/{userId}
func Put(ctx context.Context, req builtin.HttpRequest[PutRequest]) (builtin.HttpResponse[ftl.Unit], error) {
	return builtin.HttpResponse[ftl.Unit]{
		Headers: map[string][]string{"Put": {"Header from FTL"}},
		Body:    ftl.Unit{},
	}, nil
}

type DeleteRequest struct {
	UserID string `alias:"userId"`
}

type DeleteResponse struct{}

//ftl:verb
//ftl:ingress http DELETE /users/{userId}
func Delete(ctx context.Context, req builtin.HttpRequest[DeleteRequest]) (builtin.HttpResponse[ftl.Unit], error) {
	return builtin.HttpResponse[ftl.Unit]{
		Status:  200,
		Headers: map[string][]string{"Delete": {"Header from FTL"}},
		Body:    ftl.Unit{},
	}, nil
}

type HtmlRequest struct{}

//ftl:verb
//ftl:ingress http GET /html
func Html(ctx context.Context, req builtin.HttpRequest[HtmlRequest]) (builtin.HttpResponse[string], error) {
	return builtin.HttpResponse[string]{
		Headers: map[string][]string{"Content-Type": {"text/html; charset=utf-8"}},
		Body:    "<html><body><h1>HTML Page From FTL 🚀!</h1></body></html>",
	}, nil
}

//ftl:verb
//ftl:ingress http POST /bytes
func Bytes(ctx context.Context, req builtin.HttpRequest[[]byte]) (builtin.HttpResponse[[]byte], error) {
	return builtin.HttpResponse[[]byte]{
		Body: req.Body,
	}, nil
}

//ftl:verb
//ftl:ingress http GET /empty
func Empty(ctx context.Context, req builtin.HttpRequest[ftl.Unit]) (builtin.HttpResponse[ftl.Unit], error) {
	return builtin.HttpResponse[ftl.Unit]{
		Body: ftl.Unit{},
	}, nil
}

//ftl:verb
//ftl:ingress http GET /string
func String(ctx context.Context, req builtin.HttpRequest[string]) (builtin.HttpResponse[string], error) {
	return builtin.HttpResponse[string]{
		Body: req.Body,
	}, nil
}
