//ftl:module two
package two

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

//ftl:enum
type TwoEnum string

const (
	Red   TwoEnum = "Red"
	Blue  TwoEnum = "Blue"
	Green TwoEnum = "Green"
)

type User struct {
	Name string
}

type Payload[T any] struct {
	Body T
}

type UserResponse struct {
	User User
}

//ftl:verb
func Two(ctx context.Context, req Payload[string]) (Payload[string], error) {
	return Payload[string]{}, nil
}

//ftl:verb
func CallsTwo(ctx context.Context, req Payload[string]) (Payload[string], error) {
	return ftl.Call(ctx, Two, req)
}

//ftl:verb
func ReturnsUser(ctx context.Context) (ftl.Option[UserResponse], error) {
	return ftl.Some[UserResponse](UserResponse{
		User: User{
			Name: "John Doe",
		},
	}), nil
}
