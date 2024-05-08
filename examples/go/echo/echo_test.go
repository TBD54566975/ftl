package echo

import (
	"testing"

	"github.com/TBD54566975/ftl/go-runtime/ftl/ftltest"
	"github.com/stretchr/testify/assert"
)

func TestEcho(t *testing.T) {
	ctx := ftltest.Context(
		ftltest.WithConfig(defaultName, "hello"),
	)
	firstMapValue := defaultMap.Get(ctx)

	ctx = ftltest.Context(
		ftltest.WithConfig(defaultName, "world"),
	)
	secondMapValue := defaultMap.Get(ctx)

	// The mapped value needs to be different when the config value is different
	assert.Equal(t, "hello mapped", firstMapValue)
	assert.Equal(t, "world mapped", secondMapValue)
}

func TestDatabase(t *testing.T) {
	t.Setenv("FTL_POSTGRES_DSN_ECHO_ECHO", "fake")
	ctx := ftltest.Context()
	firstMapValue := dbMap.Get(ctx)

	ctx = ftltest.Context()
	secondMapValue := dbMap.Get(ctx)

	// Each context's ModuleContext initiated a different sql.DB instance with its own connection pool
	// Therefore the map should call it's mapping function twice
	assert.NotEqual(t, firstMapValue, secondMapValue)
}
