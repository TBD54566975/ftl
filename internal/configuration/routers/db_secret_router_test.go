package routers

import (
	"context"
	"net/url"
	"testing"

	"github.com/alecthomas/assert/v2"
	. "github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/configuration/routers/routerstest"
)

func TestDBSecretResolverList(t *testing.T) {
	ctx := context.Background()
	resolver := NewDatabaseSecrets(&routerstest.MockDBSecretResolverDAL{})

	rone := configuration.Ref{Module: Some("foo"), Name: "one"}
	err := resolver.Set(ctx, rone, &url.URL{Scheme: "asm", Host: rone.String()})
	assert.NoError(t, err)

	rtwo := configuration.Ref{Module: Some("foo"), Name: "two"}
	err = resolver.Set(ctx, rtwo, &url.URL{Scheme: "asm", Host: rtwo.String()})
	assert.NoError(t, err)

	entries, err := resolver.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(entries), 2)

	err = resolver.Unset(ctx, rone)
	assert.NoError(t, err)

	entries, err = resolver.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(entries), 1)

	url, err := resolver.Get(ctx, rtwo)
	assert.NoError(t, err)
	assert.Equal(t, url.String(), "asm://foo.two")
}
