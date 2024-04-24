package failing

import (
	"context"

	lib "github.com/TBD54566975/ftl/go-runtime/compile/testdata"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

var empty = ftl.Config[string]("")

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

//ftl:internal
func BadCalls(ctx context.Context, req Request) (Response, error) {
	ftl.Call(ctx, lib.OtherFunc, lib.Request{})
	ftl.Call(ctx, "failing", "failingVerb", req)
	ftl.Call(ctx, "failing", req)
	return Response{}, nil
}

//ftl:internal
func TooManyParams(ctx context.Context, req Request, req2 Request) (Response, error) {
	return Response{}, nil
}

//ftl:internal
func WrongParamOrder(first Request, second string) (Response, error) {
	return Response{}, nil
}

//ftl:internal
func UnitSecondParam(ctx context.Context, unit ftl.Unit) (Response, error) {
	return Response{}, nil
}

//ftl:internal
func NoParams() (Response, error) {
	return Response{}, nil
}

//ftl:internal
func TooManyReturn(ctx context.Context, req Request) (Response, Response, error) {
	return "", Response{}, nil
}

//ftl:internal
func NoReturn(ctx context.Context, req Request) {
}

//ftl:internal
func NoError(ctx context.Context, req Request) Response {
	return Response{}
}

//ftl:internal
func WrongResponse(ctx context.Context, req Request) (string, ftl.Unit) {
	return "", ftl.Unit{}
}

// Duplicate
//
//ftl:internal
func WrongResponse(ctx context.Context, req Request) (string, ftl.Unit) {
	return "", ftl.Unit{}
}

//ftl:internal
type BadStruct struct {
	unexported string
}
