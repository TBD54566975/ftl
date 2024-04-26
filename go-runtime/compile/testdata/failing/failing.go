package failing

import (
	"context"

	lib "github.com/TBD54566975/ftl/go-runtime/compile/testdata"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

var empty = ftl.Config[string]("")

var goodConfig = ftl.Config[string]("FTL_ENDPOINT")
var duplConfig = ftl.Config[string]("FTL_ENDPOINT")

var goodSecret = ftl.Secret[string]("FTL_ENDPOINT")
var duplSecret = ftl.Secret[string]("FTL_ENDPOINT")

var goodDB = ftl.PostgresDatabase("testDb")
var duplDB = ftl.PostgresDatabase("testDb")

type Request struct {
	BadParam error
}
type Response struct {
	AnotherBadParam uint64
}

//ftl:verb
func WrongDirective(ctx context.Context, req Request) (Response, error) {
	return Response{}, nil
}

//ftl:export
func BadCalls(ctx context.Context, req Request) (Response, error) {
	ftl.Call(ctx, lib.OtherFunc, lib.Request{})
	ftl.Call(ctx, "failing", "failingVerb", req)
	ftl.Call(ctx, "failing", req)
	return Response{}, nil
}

//ftl:export
func TooManyParams(ctx context.Context, req Request, req2 Request) (Response, error) {
	return Response{}, nil
}

//ftl:export
func WrongParamOrder(first Request, second string) (Response, error) {
	return Response{}, nil
}

//ftl:export
func UnitSecondParam(ctx context.Context, unit ftl.Unit) (Response, error) {
	return Response{}, nil
}

//ftl:export
func NoParams() (Response, error) {
	return Response{}, nil
}

//ftl:export
func TooManyReturn(ctx context.Context, req Request) (Response, Response, error) {
	return "", Response{}, nil
}

//ftl:export
func NoReturn(ctx context.Context, req Request) {
}

//ftl:export
func NoError(ctx context.Context, req Request) Response {
	return Response{}
}

//ftl:export
func WrongResponse(ctx context.Context, req Request) (string, ftl.Unit) {
	return "", ftl.Unit{}
}

// Duplicate
//
//ftl:export
func WrongResponse(ctx context.Context, req Request) (string, ftl.Unit) {
	return "", ftl.Unit{}
}

//ftl:export
type BadStruct struct {
	unexported string
}
