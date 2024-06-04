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
		ftltest.WhenVerb(Verb, func(ctx context.Context, req Request) (Response, error) {
			return Response{Output: fmt.Sprintf("fake: %s", req.Input)}, nil
		}),
		ftltest.WhenSource(Source, func(ctx context.Context) (Response, error) {
			return Response{Output: "fake"}, nil
		}),
		ftltest.WhenSink(Sink, func(ctx context.Context, req Request) error {
			knockOnEffects["sink"] = req.Input
			return nil
		}),
		ftltest.WhenEmpty(Empty, func(ctx context.Context) error {
			knockOnEffects["empty"] = "test"
			return nil
		}),
	)

	verbResp, err := ftl.Call(ctx, Verb, Request{Input: "test"})
	assert.NoError(t, err)
	assert.Equal(t, Response{Output: "fake: test"}, verbResp)

	sourceResp, err := ftl.CallSource(ctx, Source)
	assert.NoError(t, err)
	assert.Equal(t, Response{Output: "fake"}, sourceResp)

	err = ftl.CallSink(ctx, Sink, Request{Input: "testsink"})
	assert.NoError(t, err)
	assert.Equal(t, knockOnEffects["sink"], "testsink")

	err = ftl.CallEmpty(ctx, Empty)
	assert.NoError(t, err)
	assert.Equal(t, knockOnEffects["empty"], "test")
}

func TestVerbErrors(t *testing.T) {
	ctx := ftltest.Context(
		ftltest.WhenVerb(Verb, func(ctx context.Context, req Request) (Response, error) {
			return Response{}, fmt.Errorf("fake: %s", req.Input)
		}),
		ftltest.WhenSource(Source, func(ctx context.Context) (Response, error) {
			return Response{}, fmt.Errorf("fake-source")
		}),
		ftltest.WhenSink(Sink, func(ctx context.Context, req Request) error {
			return fmt.Errorf("fake: %s", req.Input)
		}),
		ftltest.WhenEmpty(Empty, func(ctx context.Context) error {
			return fmt.Errorf("fake-empty")
		}),
	)

	_, err := ftl.Call(ctx, Verb, Request{Input: "test"})
	assert.EqualError(t, err, "verbtypes.verb: fake: test")

	_, err = ftl.CallSource(ctx, Source)
	assert.EqualError(t, err, "verbtypes.source: fake-source")

	err = ftl.CallSink(ctx, Sink, Request{Input: "test-sink"})
	assert.EqualError(t, err, "verbtypes.sink: fake: test-sink")

	err = ftl.CallEmpty(ctx, Empty)
	assert.EqualError(t, err, "verbtypes.empty: fake-empty")
}
