package routers

import (
	"context"
	"net/url"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/must"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/internal/configuration"
)

func TestFileRouter(t *testing.T) {
	dir := t.TempDir()
	router := NewFileRouter[configuration.Secrets](filepath.Join(dir, "secrets.json"))
	ctx := context.Background()

	ref1 := configuration.Ref{Module: optional.Some[string]("foo"), Name: "bar"}
	url1 := must.Get(url.Parse("http://example.com"))
	ref2 := configuration.Ref{Module: optional.Some[string]("foo"), Name: "baz"}
	url2 := must.Get(url.Parse("http://example2.com"))

	err := router.Set(ctx, ref1, url1)
	assert.NoError(t, err)
	err = router.Set(ctx, ref2, url2)
	assert.NoError(t, err)

	entries, err := router.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []configuration.Entry{
		{Ref: ref1, Accessor: url1},
		{Ref: ref2, Accessor: url2},
	}, entries)

	key, err := router.Get(ctx, ref1)
	assert.NoError(t, err)
	assert.Equal(t, url1, key)

	err = router.Unset(ctx, ref1)
	assert.NoError(t, err)

	entries, err = router.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []configuration.Entry{
		{Ref: ref2, Accessor: url2},
	}, entries)

	assert.Equal(t, configuration.Secrets{}, router.Role())
}
