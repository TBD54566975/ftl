package ftltest

import (
	"context"

	"github.com/TBD54566975/ftl/go-runtime/ftl"
)

type mockFunc func(ctx context.Context, req any) (resp any, err error)

type MockProvider struct {
	mocks map[ftl.Ref]mockFunc
}

var _ = (ftl.CallOverrider)(&MockProvider{})

func NewMockProvider(ctx context.Context, mocks map[ftl.Ref]mockFunc) *MockProvider {
	provider := &MockProvider{
		mocks: mocks,
	}
	return provider
}

func (m *MockProvider) OverrideCall(ctx context.Context, ref ftl.Ref, req any) (override bool, resp any, err error) {
	mock, ok := m.mocks[ref]
	if !ok {
		return false, nil, nil
	}
	resp, err = mock(ctx, req)
	return true, resp, err
}
