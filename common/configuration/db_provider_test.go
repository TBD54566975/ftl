package configuration

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
)

func TestDBProvider(t *testing.T) {
	ctx := context.Background()
	u := URL("db://asdf")
	b := []byte("asdf")
	ref := Ref{
		Module: optional.None[string](),
		Name:   "name",
	}
	p := DBProvider{true}

	gotBytes, err := p.Load(ctx, ref, u)
	assert.Equal(t, b, gotBytes)
	assert.NoError(t, err)

	gotURL, err := p.Store(ctx, ref, b)
	assert.Equal(t, u, gotURL)
	assert.NoError(t, err)
}
