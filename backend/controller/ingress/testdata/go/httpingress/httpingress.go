package httpingress

import (
	"context"
	"fmt"

	"ftl/builtin"

	"github.com/TBD54566975/ftl/go-runtime/ftl" // Import the FTL SDK.
	lib "github.com/TBD54566975/ftl/go-runtime/schema/testdata"
)

type GetRequest struct {
	UserID string `json:"userId,omitempty"`
	PostID string `json:"postId,something,else"`
}

type Nested struct {
	GoodStuff string `json:"good_stuff"`
}

type GetResponse struct {
	Message string `json:"msg"`
	Nested  Nested `json:"nested"`
}

//ftl:enum export
type SumType interface {
	tag()
}

type A string

func (A) tag() {}

type B []string

func (B) tag() {}

//ftl:ingress http GET /users/{userId}/posts/{postId}
func Get(ctx context.Context, req builtin.HttpRequest[GetRequest]) (builtin.HttpResponse[GetResponse, string], error) {
	return builtin.HttpResponse[GetResponse, string]{
		Headers: map[string][]string{"Get": {"Header from FTL"}},
		Body: ftl.Some(GetResponse{
			Message: fmt.Sprintf("UserID: %s, PostID: %s", req.Body.UserID, req.Body.PostID),
			Nested: Nested{
				GoodStuff: "This is good stuff",
			},
		}),
	}, nil
}

type PostRequest struct {
	UserID int `json:"user_id"`
	PostID int
}

type PostResponse struct {
	Success bool `json:"success"`
}

//ftl:ingress http POST /users
func Post(ctx context.Context, req builtin.HttpRequest[PostRequest]) (builtin.HttpResponse[PostResponse, string], error) {
	return builtin.HttpResponse[PostResponse, string]{
		Status:  201,
		Headers: map[string][]string{"Post": {"Header from FTL"}},
		Body:    ftl.Some(PostResponse{Success: true}),
	}, nil
}

type PutRequest struct {
	UserID string `json:"userId"`
	PostID string `json:"postId"`
}

type PutResponse struct{}

//ftl:ingress http PUT /users/{userId}
func Put(ctx context.Context, req builtin.HttpRequest[PutRequest]) (builtin.HttpResponse[builtin.Empty, string], error) {
	return builtin.HttpResponse[builtin.Empty, string]{
		Headers: map[string][]string{"Put": {"Header from FTL"}},
		Body:    ftl.Some(builtin.Empty{}),
	}, nil
}

type DeleteRequest struct {
	UserID string `json:"userId"`
}

type DeleteResponse struct{}

//ftl:ingress http DELETE /users/{userId}
func Delete(ctx context.Context, req builtin.HttpRequest[DeleteRequest]) (builtin.HttpResponse[builtin.Empty, string], error) {
	return builtin.HttpResponse[builtin.Empty, string]{
		Status:  200,
		Headers: map[string][]string{"Delete": {"Header from FTL"}},
		Body:    ftl.Some(builtin.Empty{}),
	}, nil
}

type QueryParamRequest struct {
	Foo ftl.Option[string] `json:"foo"`
}

//ftl:ingress http GET /queryparams
func Query(ctx context.Context, req builtin.HttpRequest[QueryParamRequest]) (builtin.HttpResponse[string, string], error) {
	return builtin.HttpResponse[string, string]{
		Body: ftl.Some(req.Body.Foo.Default("No value")),
	}, nil
}

type HtmlRequest struct{}

//ftl:ingress http GET /html
func Html(ctx context.Context, req builtin.HttpRequest[HtmlRequest]) (builtin.HttpResponse[string, string], error) {
	return builtin.HttpResponse[string, string]{
		Headers: map[string][]string{"Content-Type": {"text/html; charset=utf-8"}},
		Body:    ftl.Some("<html><body><h1>HTML Page From FTL 🚀!</h1></body></html>"),
	}, nil
}

//ftl:ingress http POST /bytes
func Bytes(ctx context.Context, req builtin.HttpRequest[[]byte]) (builtin.HttpResponse[[]byte, string], error) {
	return builtin.HttpResponse[[]byte, string]{Body: ftl.Some(req.Body)}, nil
}

//ftl:ingress http GET /empty
func Empty(ctx context.Context, req builtin.HttpRequest[ftl.Unit]) (builtin.HttpResponse[ftl.Unit, string], error) {
	return builtin.HttpResponse[ftl.Unit, string]{Body: ftl.Some(ftl.Unit{})}, nil
}

//ftl:ingress http GET /string
func String(ctx context.Context, req builtin.HttpRequest[string]) (builtin.HttpResponse[string, string], error) {
	return builtin.HttpResponse[string, string]{Body: ftl.Some(req.Body)}, nil
}

//ftl:ingress http GET /int
func Int(ctx context.Context, req builtin.HttpRequest[int]) (builtin.HttpResponse[int, string], error) {
	return builtin.HttpResponse[int, string]{Body: ftl.Some(req.Body)}, nil
}

//ftl:ingress http GET /float
func Float(ctx context.Context, req builtin.HttpRequest[float64]) (builtin.HttpResponse[float64, string], error) {
	return builtin.HttpResponse[float64, string]{Body: ftl.Some(req.Body)}, nil
}

//ftl:ingress http GET /bool
func Bool(ctx context.Context, req builtin.HttpRequest[bool]) (builtin.HttpResponse[bool, string], error) {
	return builtin.HttpResponse[bool, string]{Body: ftl.Some(req.Body)}, nil
}

//ftl:ingress http GET /error
func Error(ctx context.Context, req builtin.HttpRequest[ftl.Unit]) (builtin.HttpResponse[ftl.Unit, string], error) {
	return builtin.HttpResponse[ftl.Unit, string]{
		Status: 500,
		Error:  ftl.Some("Error from FTL"),
	}, nil
}

//ftl:ingress http GET /array/string
func ArrayString(ctx context.Context, req builtin.HttpRequest[[]string]) (builtin.HttpResponse[[]string, string], error) {
	return builtin.HttpResponse[[]string, string]{
		Body: ftl.Some(req.Body),
	}, nil
}

type ArrayType struct {
	Item string `json:"item"`
}

//ftl:ingress http POST /array/data
func ArrayData(ctx context.Context, req builtin.HttpRequest[[]ArrayType]) (builtin.HttpResponse[[]ArrayType, string], error) {
	return builtin.HttpResponse[[]ArrayType, string]{
		Body: ftl.Some(req.Body),
	}, nil
}

//ftl:ingress http GET /typeenum
func TypeEnum(ctx context.Context, req builtin.HttpRequest[SumType]) (builtin.HttpResponse[SumType, string], error) {
	return builtin.HttpResponse[SumType, string]{Body: ftl.Some(req.Body)}, nil
}

// tests both supported patterns for aliasing an external type

type NewTypeAlias lib.NonFTLType

//ftl:ingress http GET /external
func External(ctx context.Context, req builtin.HttpRequest[NewTypeAlias]) (builtin.HttpResponse[NewTypeAlias, string], error) {
	return builtin.HttpResponse[NewTypeAlias, string]{Body: ftl.Some(req.Body)}, nil
}

type DirectTypeAlias = lib.NonFTLType

//ftl:ingress http GET /external2
func External2(ctx context.Context, req builtin.HttpRequest[DirectTypeAlias]) (builtin.HttpResponse[DirectTypeAlias, string], error) {
	return builtin.HttpResponse[DirectTypeAlias, string]{Body: ftl.Some(req.Body)}, nil
}
