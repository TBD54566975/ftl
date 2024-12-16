package ingress

import (
	"context"
	"fmt"

	"ftl/builtin"

	"github.com/block/ftl/go-runtime/ftl" // Import the FTL SDK.
)

type GetRequest struct {
	UserID int `json:"userId,omitempty"`
	PostID int `json:"postId,omitempty"`
}

type GetResponse struct {
	Message string `json:"message,omitempty"`
}

//ftl:ingress http GET /users/{userId}/posts/{postId}
func Get(ctx context.Context, req builtin.HttpRequest[ftl.Unit, GetRequest, ftl.Unit]) (builtin.HttpResponse[GetResponse, string], error) {
	return builtin.HttpResponse[GetResponse, string]{
		Headers: map[string][]string{"Get": {"Header from FTL"}},
		Body: ftl.Some(GetResponse{
			Message: fmt.Sprintf("UserID: %d, PostID: %d", req.PathParameters.UserID, req.PathParameters.PostID),
		}),
	}, nil
}
