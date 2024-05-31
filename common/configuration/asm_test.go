package configuration

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/TBD54566975/ftl/internal/log"

	"github.com/alecthomas/assert/v2"
	. "github.com/alecthomas/types/optional"
)

func localstack() ASM[Secrets] {
	return ASM[Secrets]{
		AccessKeyID:     "test",
		SecretAccessKey: "test",
		Region:          "us-west-2",
		Endpoint:        Some("http://localhost:4566"),
	}
}

func TestASMWorkflow(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	asm := localstack()
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

// Suggest not running this against a real AWS account (especially in CI) due to the cost. Maybe costs a few $.
func TestASMPagination(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	asm := localstack()

	// Create 210 secrets, so we paginate at least twice.
	for i := range 210 {
		ref := NewRef("foo", fmt.Sprintf("bar%03d", i))
		_, err := asm.Store(ctx, ref, []byte(fmt.Sprintf("hunter%03d", i)))
		assert.NoError(t, err)
	}

	items, err := asm.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(items), 210)

	// Check each secret.
	sort.Slice(items, func(i, j int) bool {
		return items[i].Ref.Name < items[j].Ref.Name
	})
	for i, item := range items {
		assert.Equal(t, item.Ref.Name, fmt.Sprintf("bar%03d", i))

		// Just to save on requests, skip by 10
		if i%10 != 0 {
			continue
		}
		secret, err := asm.Load(ctx, item.Ref, item.Accessor)
		assert.NoError(t, err)
		assert.Equal(t, secret, []byte(fmt.Sprintf("hunter%03d", i)))
	}

	// Delete them
	for i := range 210 {
		ref := NewRef("foo", fmt.Sprintf("bar%03d", i))
		err := asm.Delete(ctx, ref)
		assert.NoError(t, err)
	}

	// Make sure they are all gone
	items, err = asm.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(items), 0)
}
