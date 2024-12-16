package two

import (
	"context"

	"github.com/block/ftl/go-runtime/ftl"
	lib "github.com/block/ftl/go-runtime/schema/testdata"
	libbackoff "github.com/jpillora/backoff"

	"ftl/builtin"
)

type FooConfig struct {
	ftl.DefaultPostgresDatabaseConfig
}

func (FooConfig) Name() string { return "foo" }

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
func Two(ctx context.Context, req Payload[string], handle ftl.DatabaseHandle[FooConfig]) (Payload[string], error) {
	return Payload[string]{}, nil
}

//ftl:verb export
func Three(ctx context.Context, req Payload[string]) (Payload[string], error) {
	return Payload[string]{}, nil
}

//ftl:verb export
func CallsTwo(ctx context.Context, req Payload[string], two TwoClient) (Payload[string], error) {
	return two(ctx, req)
}

//ftl:verb export
func CallsTwoAndThree(ctx context.Context, req Payload[string], two TwoClient, three ThreeClient) (Payload[string], error) {
	err := transitiveVerbCall(ctx, req, two, three)
	return Payload[string]{}, err
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

type TransitiveAlias lib.NonFTLType

//ftl:typealias
type BackoffAlias libbackoff.Backoff

func transitiveVerbCall(ctx context.Context, req Payload[string], two TwoClient, three ThreeClient) error {
	_, err := two(ctx, req)
	if err != nil {
		return err
	}
	err = superTransitiveVerbCall(ctx, req, three)
	return err
}

func superTransitiveVerbCall(ctx context.Context, req Payload[string], three ThreeClient) error {
	_, err := three(ctx, req)
	return err
}

type PaymentState string

type PayinState PaymentState

const (
	PayinPending PayinState = "PAYIN_PENDING"
)

type PayoutState PaymentState

const (
	PayoutPending PayoutState = "PAYOUT_PENDING"
)

//ftl:data
type Payment struct {
	In  PayinState
	Out PayoutState
}

type PostRequest struct {
	UserID int
	PostID int
}

type PostResponse struct {
	Success bool
}

//ftl:ingress http POST /users
//ftl:encoding lenient
func Ingress(ctx context.Context, req builtin.HttpRequest[PostRequest, ftl.Unit, ftl.Unit]) (builtin.HttpResponse[PostResponse, string], error) {
	return builtin.HttpResponse[PostResponse, string]{
		Status:  201,
		Headers: map[string][]string{"Post": {"Header from FTL"}},
		Body:    ftl.Some(PostResponse{Success: true}),
	}, nil
}
