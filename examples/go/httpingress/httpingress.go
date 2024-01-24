//ftl:module httpingress
package httpingress

import (
	"context"
	"fmt"

	"ftl/builtin"

	ftl "github.com/TBD54566975/ftl/go-runtime/sdk"
)

type GetRequest struct {
	UserID string `json:"userId"`
	PostID string `json:"postId"`
}

type GetResponse struct {
	Message string `json:"message"`
}

//ftl:verb
//ftl:ingress http GET /http/users/{userID}/posts/{postID}
func Get(ctx context.Context, req builtin.HttpRequest[GetRequest]) (builtin.HttpResponse[GetResponse], error) {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("Path: %s", req.Path)
	logger.Infof("Method: %s", req.Method)
	logger.Infof("Query: %s", req.Query)
	logger.Infof("Body: %s", req.Body)
	logger.Infof("Headers: %s", req.Headers)
	return builtin.HttpResponse[GetResponse]{
		Status:  200,
		Headers: map[string][]string{"Get": {"Header from FTL"}},
		Body:    GetResponse{Message: fmt.Sprintf("UserID, %s : PostID %s", req.Body.UserID, req.Body.PostID)},
	}, nil
}

type PostRequest struct {
	UserID string `json:"userId"`
	PostID string `json:"postId"`
}

type PostResponse struct{}

//ftl:verb
//ftl:ingress http POST /http/users
func Post(ctx context.Context, req builtin.HttpRequest[PostRequest]) (builtin.HttpResponse[PostResponse], error) {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("Path: %s", req.Path)
	logger.Infof("Method: %s", req.Method)
	logger.Infof("Query: %s", req.Query)
	logger.Infof("Body: %s", req.Body)
	logger.Infof("Headers: %s", req.Headers)
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
//ftl:ingress http PUT /http/users/{userID}
func Put(ctx context.Context, req builtin.HttpRequest[PutRequest]) (builtin.HttpResponse[PutResponse], error) {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("Path: %s", req.Path)
	logger.Infof("Method: %s", req.Method)
	logger.Infof("Query: %s", req.Query)
	logger.Infof("Body: %s", req.Body)
	logger.Infof("Headers: %s", req.Headers)
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
//ftl:ingress http DELETE /http/users/{userID}
func Delete(ctx context.Context, req builtin.HttpRequest[DeleteRequest]) (builtin.HttpResponse[DeleteResponse], error) {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("Path: %s", req.Path)
	logger.Infof("Method: %s", req.Method)
	logger.Infof("Query: %s", req.Query)
	logger.Infof("Body: %s", req.Body)
	logger.Infof("Headers: %s", req.Headers)
	return builtin.HttpResponse[DeleteResponse]{
		Status:  200,
		Headers: map[string][]string{"Put": {"Header from FTL"}},
		Body:    DeleteResponse{},
	}, nil
}
