//ftl:module echo
package echo

import (
	"context"
	"fmt"

	"ftl/builtin"
)

type GetRequest struct {
	UserID string `json:"userId"`
	PostID string `json:"postId"`
}

type GetResponse struct {
	Message string `json:"message"`
}

//ftl:verb
//ftl:ingress http GET /echo/users/{userID}/posts/{postID}
func Get(ctx context.Context, req builtin.HttpRequest[GetRequest]) (builtin.HttpResponse[GetResponse], error) {
	return builtin.HttpResponse[GetResponse]{
		Status:  200,
		Headers: map[string][]string{"Get": {"Header from FTL"}},
		Body:    GetResponse{Message: fmt.Sprintf("UserID: %s, PostID: %s", req.Body.UserID, req.Body.PostID)},
	}, nil
}

type PostRequest struct {
	UserID string `json:"userId" alias:"id"`
	PostID string `json:"postId"`
}

type PostResponse struct{}

//ftl:verb
//ftl:ingress http POST /echo/users
func Post(ctx context.Context, req builtin.HttpRequest[PostRequest]) (builtin.HttpResponse[PostResponse], error) {
	return builtin.HttpResponse[PostResponse]{
		Status:  201,
		Headers: map[string][]string{"Post": {"Header from FTL"}},
		Body:    PostResponse{},
	}, nil
}

type PutRequest struct {
	UserID string `json:"userId"`
	PostID string `json:"postId"`
}

type PutResponse struct{}

//ftl:verb
//ftl:ingress http PUT /echo/users/{userID}
func Put(ctx context.Context, req builtin.HttpRequest[PutRequest]) (builtin.HttpResponse[PutResponse], error) {
	return builtin.HttpResponse[PutResponse]{
		Status:  200,
		Headers: map[string][]string{"Put": {"Header from FTL"}},
		Body:    PutResponse{},
	}, nil
}

type DeleteRequest struct {
	UserID string `json:"userId"`
}

type DeleteResponse struct{}

//ftl:verb
//ftl:ingress http DELETE /echo/users/{userID}
func Delete(ctx context.Context, req builtin.HttpRequest[DeleteRequest]) (builtin.HttpResponse[DeleteResponse], error) {
	return builtin.HttpResponse[DeleteResponse]{
		Status:  200,
		Headers: map[string][]string{"Put": {"Header from FTL"}},
		Body:    DeleteResponse{},
	}, nil
}
