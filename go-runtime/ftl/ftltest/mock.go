package ftltest

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

type mockFunc func(ctx context.Context, req any) (resp any, err error)

// mockProvider keeps a mapping of verb references to mock functions.
//
// It implements the CallOverrider interface to intercept calls with the mock functions.
type mockProvider struct {
	mocks map[ftl.Ref]mockFunc
}

var _ = (ftl.CallOverrider)(&mockProvider{})

func newMockProvider() *mockProvider {
	provider := &mockProvider{
		mocks: map[ftl.Ref]mockFunc{},
	}
	return provider
}

func (m *mockProvider) OverrideCall(ctx context.Context, ref ftl.Ref, req any) (override bool, resp any, err error) {
	mock, ok := m.mocks[ref]
	if !ok {
		return false, nil, nil
	}
	resp, err = mock(ctx, req)
	return true, resp, err
}
