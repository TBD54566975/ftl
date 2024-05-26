package configuration

import (
	"context"
	"testing"

	"github.com/TBD54566975/ftl/internal/log"

	"github.com/alecthomas/assert/v2"
	. "github.com/alecthomas/types/optional"
)

func TestAWSSecretsBasics(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())

	// Localstack!
	asm := AWSSecrets[Secrets]{
		AccessKeyId:     "test",
		SecretAccessKey: "test",
		Region:          "us-west-2",
		Endpoint:        Some("http://localhost:4566"),
	}
	url := URL("asm://foo.bar")
	ref := Ref{Module: Some("foo"), Name: "bar"}
	var mySecret = []byte("my secret")

	err := asm.Set(ctx, ref, url)
	assert.NoError(t, err)

	url1, err := asm.Get(ctx, ref)
	assert.NoError(t, err)
	assert.Equal(t, url, url1)

	items, err := asm.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, items, []Entry{})

	url2, err := asm.Store(ctx, ref, mySecret)
	assert.NoError(t, err)
	assert.Equal(t, url, url2)

	items, err = asm.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, items, []Entry{{Ref: ref, Accessor: url}})

	item, err := asm.Load(ctx, ref, url)
	assert.NoError(t, err)
	assert.Equal(t, item, mySecret)

	// Store a second time to make sure it is updating
	var mySecret2 = []byte("hunter1")
	url3, err := asm.Store(ctx, ref, mySecret2)
	assert.NoError(t, err)
	assert.Equal(t, url, url3)

	item, err = asm.Load(ctx, ref, url)
	assert.NoError(t, err)
	assert.Equal(t, item, mySecret2)

	err = asm.Delete(ctx, ref)
	assert.NoError(t, err)

	items, err = asm.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, items, []Entry{})

	_, err = asm.Load(ctx, ref, url)
	assert.Error(t, err)
}
