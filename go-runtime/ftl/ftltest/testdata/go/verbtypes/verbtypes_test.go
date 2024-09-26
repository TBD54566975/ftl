package verbtypes

import (
	"context"
	"fmt"
	"testing"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/go-runtime/ftl/ftltest"
	"github.com/alecthomas/assert/v2"
)

func TestVerbs(t *testing.T) {
	knockOnEffects := map[string]string{}

	ctx := ftltest.Context(
		ftltest.WhenVerb[VerbClient](func(ctx context.Context, req Request) (Response, error) {
			return Response{Output: fmt.Sprintf("fake: %s", req.Input)}, nil
		}),
		ftltest.WhenSource[SourceClient](func(ctx context.Context) (Response, error) {
			return Response{Output: "fake"}, nil
		}),
		ftltest.WhenSink[SinkClient](func(ctx context.Context, req Request) error {
			knockOnEffects["sink"] = req.Input
			return nil
		}),
		ftltest.WhenEmpty[EmptyClient](func(ctx context.Context) error {
			knockOnEffects["empty"] = "test"
			return nil
		}),
	)

	verbResp, err := ftltest.Call[VerbClient, Request, Response](ctx, Request{Input: "test"})
	assert.NoError(t, err)
	assert.Equal(t, Response{Output: "fake: test"}, verbResp)

	sourceResp, err := ftltest.CallSource[SourceClient, Response](ctx)
	assert.NoError(t, err)
	assert.Equal(t, Response{Output: "fake"}, sourceResp)

	err = ftltest.CallSink[SinkClient](ctx, Request{Input: "testsink"})
	assert.NoError(t, err)
	assert.Equal(t, knockOnEffects["sink"], "testsink")

	err = ftltest.CallEmpty[EmptyClient](ctx)
	assert.NoError(t, err)
	assert.Equal(t, knockOnEffects["empty"], "test")

	// TODO: remove after refactor
	verbResp, err = ftl.Call(ctx, Verb, Request{Input: "test"})
	assert.NoError(t, err)
	assert.Equal(t, Response{Output: "fake: test"}, verbResp)

	sourceResp, err = ftl.CallSource(ctx, Source)
	assert.NoError(t, err)
	assert.Equal(t, Response{Output: "fake"}, sourceResp)

	err = ftl.CallSink(ctx, Sink, Request{Input: "testsink"})
	assert.NoError(t, err)
	assert.Equal(t, knockOnEffects["sink"], "testsink")

	err = ftl.CallEmpty(ctx, Empty)
	assert.NoError(t, err)
	assert.Equal(t, knockOnEffects["empty"], "test")
}

func TestContextExtension(t *testing.T) {
	ctx1 := ftltest.Context(
		ftltest.WhenSource[SourceClient](func(ctx context.Context) (Response, error) {
			return Response{Output: "fake"}, nil
		}),
	)

	ctx2 := ftltest.SubContext(
		ctx1,
		ftltest.WhenSource[SourceClient](func(ctx context.Context) (Response, error) {
			return Response{Output: "another fake"}, nil
		}),
	)

	sourceResp, err := ftltest.CallSource[SourceClient, Response](ctx1)
	assert.NoError(t, err)
	assert.Equal(t, Response{Output: "fake"}, sourceResp)

	sourceResp, err = ftltest.CallSource[SourceClient, Response](ctx2)
	assert.NoError(t, err)
	assert.Equal(t, Response{Output: "another fake"}, sourceResp)
}

func TestVerbErrors(t *testing.T) {
	ctx := ftltest.Context(
		ftltest.WhenVerb[VerbClient](func(ctx context.Context, req Request) (Response, error) {
			return Response{}, fmt.Errorf("fake: %s", req.Input)
		}),
		ftltest.WhenSource[SourceClient](func(ctx context.Context) (Response, error) {
			return Response{}, fmt.Errorf("fake-source")
		}),
		ftltest.WhenSink[SinkClient](func(ctx context.Context, req Request) error {
			return fmt.Errorf("fake: %s", req.Input)
		}),
		ftltest.WhenEmpty[EmptyClient](func(ctx context.Context) error {
			return fmt.Errorf("fake-empty")
		}),
	)

	_, err := ftltest.Call[VerbClient, Request, Response](ctx, Request{Input: "test"})
	assert.EqualError(t, err, "verbtypes.verb: fake: test")

	_, err = ftltest.CallSource[SourceClient, Response](ctx)
	assert.EqualError(t, err, "verbtypes.source: fake-source")

	err = ftltest.CallSink[SinkClient](ctx, Request{Input: "test-sink"})
	assert.EqualError(t, err, "verbtypes.sink: fake: test-sink")

	err = ftltest.CallEmpty[EmptyClient](ctx)
	assert.EqualError(t, err, "verbtypes.empty: fake-empty")

	// TODO: remove after refactor
	_, err = ftl.Call(ctx, Verb, Request{Input: "test"})
	assert.EqualError(t, err, "verbtypes.verb: fake: test")

	_, err = ftl.CallSource(ctx, Source)
	assert.EqualError(t, err, "verbtypes.source: fake-source")

	err = ftl.CallSink(ctx, Sink, Request{Input: "test-sink"})
	assert.EqualError(t, err, "verbtypes.sink: fake: test-sink")

	err = ftl.CallEmpty(ctx, Empty)
	assert.EqualError(t, err, "verbtypes.empty: fake-empty")
}

func TestTransitiveVerbMock(t *testing.T) {
	ctx := ftltest.Context(
		ftltest.WithCallsAllowedWithinModule(),
		ftltest.WhenVerb[CalleeVerbClient](func(ctx context.Context, req Request) (Response, error) {
			return Response{Output: fmt.Sprintf("mocked: %s", req.Input)}, nil
		}),
	)

	verbResp, err := ftltest.Call[CallerVerbClient, Request, Response](ctx, Request{Input: "test"})
	assert.NoError(t, err)
	assert.Equal(t, Response{Output: "mocked: test"}, verbResp)
}
