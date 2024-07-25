package mapper

import (
	"context"
	"testing"

	"github.com/TBD54566975/ftl/go-runtime/ftl/ftltest"
	"github.com/alecthomas/assert/v2"
)

func TestGet(t *testing.T) {
	ctx := ftltest.Context(ftltest.WithMapsAllowed())
	got := m.Get(ctx)
	assert.Equal(t, got, 9)
}

func TestPanicsWithoutExplicitlyAllowingMaps(t *testing.T) {
	ctx := ftltest.Context()
	assert.Panics(t, func() {
		m.Get(ctx)
	})
}

func TestMockGet(t *testing.T) {
	mockOut := 123
	ctx := ftltest.Context(ftltest.WhenMap(m, func(ctx context.Context) (int, error) {
		return mockOut, nil
	}))
	got := m.Get(ctx)
	assert.Equal(t, got, mockOut)
}
