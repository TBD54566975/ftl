package two

import (
	"context"

	lib "github.com/TBD54566975/ftl/go-runtime/compile/testdata"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

//ftl:enum export
type TwoEnum string

const (
	Red   TwoEnum = "Red"
	Blue  TwoEnum = "Blue"
	Green TwoEnum = "Green"
)

//ftl:enum export
type TypeEnum interface{ typeEnum() }

type Scalar string

func (Scalar) typeEnum() {}

type List []string

func (List) typeEnum() {}

//ftl:data export
type Exported struct {
}

func (Exported) typeEnum() {}

type WithoutDirective struct{}

func (WithoutDirective) typeEnum() {}

type User struct {
	Name string
}

type Payload[T any] struct {
	Body T
}

type UserResponse struct {
	User User
}

//ftl:verb export
func Two(ctx context.Context, req Payload[string]) (Payload[string], error) {
	return Payload[string]{}, nil
}

//ftl:verb export
func CallsTwo(ctx context.Context, req Payload[string]) (Payload[string], error) {
	return ftl.Call(ctx, Two, req)
}

//ftl:verb export
func ReturnsUser(ctx context.Context) (UserResponse, error) {
	return UserResponse{
		User: User{
			Name: "John Doe",
		},
	}, nil
}

//ftl:data
type NonFTLField struct {
	ExplicitType    ExplicitAliasType
	ExplicitAlias   ExplicitAliasAlias
	TransitiveType  TransitiveAliasType
	TransitiveAlias TransitiveAliasAlias
}

//ftl:typealias
//ftl:typemap kotlin "com.foo.bar.NonFTLType"
type ExplicitAliasType lib.NonFTLType

//ftl:typealias
//ftl:typemap kotlin "com.foo.bar.NonFTLType"
type ExplicitAliasAlias = lib.NonFTLType

type TransitiveAliasType lib.NonFTLType

type TransitiveAliasAlias = lib.NonFTLType
