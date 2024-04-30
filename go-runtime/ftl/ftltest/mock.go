package ftltest

import (
	"context"
	"fmt"

	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type mockFunc func(ctx context.Context, req any) (resp any, err error)

// mockVerbProvider keeps a mapping of verb references to mock functions.
//
// It implements the CallOverrider interface to intercept calls with the mock functions.
type mockVerbProvider struct {
	mocks map[ftl.Ref]mockFunc
}

var _ = (ftl.CallOverrider)(&mockVerbProvider{})

func newMockVerbProvider() *mockVerbProvider {
	provider := &mockVerbProvider{
		mocks: map[ftl.Ref]mockFunc{},
	}
	return provider
}

func (m *mockVerbProvider) OverrideCall(ctx context.Context, ref ftl.Ref, req any) (override bool, resp any, err error) {
	mock, ok := m.mocks[ref]
	if ok {
		resp, err = mock(ctx, req)
		return true, resp, err
	}
	if rpc.IsClientAvailableInContext[ftlv1connect.VerbServiceClient](ctx) {
		return false, nil, nil
	}
	// Return a clean error for testing because we know the client is not available to make real calls
	return false, nil, fmt.Errorf("no mock found")
}
