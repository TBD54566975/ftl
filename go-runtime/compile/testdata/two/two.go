package two

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

//ftl:export
type TwoEnum string

const (
	Red   TwoEnum = "Red"
	Blue  TwoEnum = "Blue"
	Green TwoEnum = "Green"
)

//ftl:export
type Exported struct {
}

type User struct {
	Name string
}

type Payload[T any] struct {
	Body T
}

type UserResponse struct {
	User User
}

//ftl:export
func Two(ctx context.Context, req Payload[string]) (Payload[string], error) {
	return Payload[string]{}, nil
}

//ftl:export
func CallsTwo(ctx context.Context, req Payload[string]) (Payload[string], error) {
	return ftl.Call(ctx, Two, req)
}

//ftl:export
func ReturnsUser(ctx context.Context) (UserResponse, error) {
	return UserResponse{
		User: User{
			Name: "John Doe",
		},
	}, nil
}
