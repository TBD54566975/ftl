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
func Get(ctx context.Context, req builtin.HttpRequest[ftl.Unit, GetRequest, ftl.Unit]) (builtin.HttpResponse[GetResponse, string], error) {
	return builtin.HttpResponse[GetResponse, string]{
		Headers: map[string][]string{"Get": {"Header from FTL"}},
		Body: ftl.Some(GetResponse{
			Message: fmt.Sprintf("UserID: %s, PostID: %s", req.PathParameters.UserID, req.PathParameters.PostID),
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
func Post(ctx context.Context, req builtin.HttpRequest[PostRequest, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[PostResponse, string], error) {
	return builtin.HttpResponse[PostResponse, string]{
		Status:  201,
		Headers: map[string][]string{"Post": {"Header from FTL"}},
		Body:    ftl.Some(PostResponse{Success: true}),
	}, nil
}

type PutRequest struct {
	PostID string `json:"postId"`
}

type PutResponse struct{}

//ftl:ingress http PUT /users/{userId}
func Put(ctx context.Context, req builtin.HttpRequest[PutRequest, string, ftl.Unit]) (builtin.HttpResponse[PutResponse, string], error) {
	return builtin.HttpResponse[PutResponse, string]{
		Headers: map[string][]string{"Put": {"Header from FTL"}},
		Body:    ftl.Some(PutResponse{}),
	}, nil
}

type DeleteRequest struct {
	UserID string `json:"userId"`
}

type DeleteResponse struct{}

//ftl:ingress http DELETE /users/{userId}
func Delete(ctx context.Context, req builtin.HttpRequest[ftl.Unit, DeleteRequest, ftl.Unit]) (builtin.HttpResponse[builtin.Empty, string], error) {
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
func Query(ctx context.Context, req builtin.HttpRequest[ftl.Unit, ftl.Unit, QueryParamRequest]) (builtin.HttpResponse[string, string], error) {
	return builtin.HttpResponse[string, string]{
		Body: ftl.Some(req.Query.Foo.Default("No value")),
	}, nil
}

//ftl:ingress http GET /html
func Html(ctx context.Context, req builtin.HttpRequest[ftl.Unit, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[string, string], error) {
	return builtin.HttpResponse[string, string]{
		Headers: map[string][]string{"Content-Type": {"text/html; charset=utf-8"}},
		Body:    ftl.Some("<html><body><h1>HTML Page From FTL 🚀!</h1></body></html>"),
	}, nil
}

//ftl:ingress http POST /bytes
func Bytes(ctx context.Context, req builtin.HttpRequest[[]byte, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[[]byte, string], error) {
	return builtin.HttpResponse[[]byte, string]{Body: ftl.Some(req.Body)}, nil
}

//ftl:ingress http GET /empty
func Empty(ctx context.Context, req builtin.HttpRequest[ftl.Unit, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[ftl.Unit, string], error) {
	return builtin.HttpResponse[ftl.Unit, string]{Body: ftl.Some(ftl.Unit{})}, nil
}

//ftl:ingress http POST /string
func String(ctx context.Context, req builtin.HttpRequest[string, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[string, string], error) {
	return builtin.HttpResponse[string, string]{Body: ftl.Some(req.Body)}, nil
}

//ftl:ingress http POST /int
func Int(ctx context.Context, req builtin.HttpRequest[int, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[int, string], error) {
	return builtin.HttpResponse[int, string]{Body: ftl.Some(req.Body)}, nil
}

//ftl:ingress http POST /float
func Float(ctx context.Context, req builtin.HttpRequest[float64, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[float64, string], error) {
	return builtin.HttpResponse[float64, string]{Body: ftl.Some(req.Body)}, nil
}

//ftl:ingress http POST /bool
func Bool(ctx context.Context, req builtin.HttpRequest[bool, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[bool, string], error) {
	return builtin.HttpResponse[bool, string]{Body: ftl.Some(req.Body)}, nil
}

//ftl:ingress http GET /error
func Error(ctx context.Context, req builtin.HttpRequest[ftl.Unit, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[ftl.Unit, string], error) {
	return builtin.HttpResponse[ftl.Unit, string]{
		Status: 500,
		Error:  ftl.Some("Error from FTL"),
	}, nil
}

//ftl:ingress http POST /array/string
func ArrayString(ctx context.Context, req builtin.HttpRequest[[]string, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[[]string, string], error) {
	return builtin.HttpResponse[[]string, string]{
		Body: ftl.Some(req.Body),
	}, nil
}

type ArrayType struct {
	Item string `json:"item"`
}

//ftl:ingress http POST /array/data
func ArrayData(ctx context.Context, req builtin.HttpRequest[[]ArrayType, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[[]ArrayType, string], error) {
	return builtin.HttpResponse[[]ArrayType, string]{
		Body: ftl.Some(req.Body),
	}, nil
}

//ftl:ingress http POST /typeenum
func TypeEnum(ctx context.Context, req builtin.HttpRequest[SumType, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[SumType, string], error) {
	return builtin.HttpResponse[SumType, string]{Body: ftl.Some(req.Body)}, nil
}

// tests both supported patterns for aliasing an external type

type NewTypeAlias lib.NonFTLType

//ftl:ingress http POST /external
func External(ctx context.Context, req builtin.HttpRequest[NewTypeAlias, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[NewTypeAlias, string], error) {
	return builtin.HttpResponse[NewTypeAlias, string]{Body: ftl.Some(req.Body)}, nil
}

type DirectTypeAlias = lib.NonFTLType

//ftl:ingress http POST /external2
func External2(ctx context.Context, req builtin.HttpRequest[DirectTypeAlias, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[DirectTypeAlias, string], error) {
	return builtin.HttpResponse[DirectTypeAlias, string]{Body: ftl.Some(req.Body)}, nil
}

//ftl:ingress http POST /lenient
//ftl:encoding lenient
func Lenient(ctx context.Context, req builtin.HttpRequest[PostRequest, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[PostResponse, string], error) {
	return builtin.HttpResponse[PostResponse, string]{
		Status:  201,
		Headers: map[string][]string{"Post": {"Header from FTL"}},
		Body:    ftl.Some(PostResponse{Success: true}),
	}, nil
}
