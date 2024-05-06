package verbtypes

import (
	"context"
	// Import the FTL SDK.
)

// verbtypes is a simple module that has each type of verb (verb, source, sink and empty)

type Request struct {
	Input string `json:"input"`
}

type Response struct {
	Output string `json:"output"`
}

//ftl:verb
func Verb(ctx context.Context, req Request) (Response, error) {
	return Response{Output: req.Input}, nil
}

//ftl:verb
func Source(ctx context.Context) (Response, error) {
	return Response{Output: "source"}, nil
}

//ftl:verb
func Sink(ctx context.Context, req Request) error {
	return nil
}

//ftl:verb
func Empty(ctx context.Context) error {
	return nil
}
