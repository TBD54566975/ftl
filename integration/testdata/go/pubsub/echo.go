package echo

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	// Import the FTL SDK.
)

type EchoRequest struct{}
type EchoResponse struct {
	Name string
}

//ftl:internal
func Echo(ctx context.Context, req EchoRequest) (EchoResponse, error) {
	err := ftl.CallSink(ctx, Sink, SinkRequest{})
	if err != nil {
		return EchoResponse{}, err
	}
	resp, err := ftl.CallSource(ctx, Source)
	if err != nil {
		return EchoResponse{}, err
	}

	name := resp.Name
	return EchoResponse{
		Name: name,
	}, nil
}

type SourceResponse struct {
	Name string
}

//ftl:internal
func Source(ctx context.Context) (SourceResponse, error) {
	return SourceResponse{
		Name: "source",
	}, nil
}

type SinkRequest struct{}

//ftl:internal
func Sink(ctx context.Context, req SinkRequest) error {
	return nil
}
