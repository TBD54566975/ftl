package mapper

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/TBD54566975/ftl/go-runtime/ftl/ftltest"
)

func TestGet(t *testing.T) {
	ctx := ftltest.Context(t, ftltest.WithMapsAllowed())
	got := m.Get(ctx)
	assert.Equal(t, got, 9)
}

func TestPanicsWithoutExplicitlyAllowingMaps(t *testing.T) {
	ctx := ftltest.Context(t)
	assert.Panics(t, func() {
		m.Get(ctx)
	})
}

func TestMockGet(t *testing.T) {
	mockOut := 123
	ctx := ftltest.Context(t, ftltest.WhenMap(m, func(ctx context.Context) (int, error) {
		return mockOut, nil
	}))
	got := m.Get(ctx)
	assert.Equal(t, got, mockOut)
}
