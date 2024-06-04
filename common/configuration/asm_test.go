//go:build integration

package configuration

import (
	"context"
	"fmt"
	"sort"
	"testing"

	"github.com/TBD54566975/ftl/internal/log"

	"github.com/alecthomas/assert/v2"
	. "github.com/alecthomas/types/optional"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func localstack(ctx context.Context, t *testing.T) ASM {
	t.Helper()
	cc := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test", "test", ""))
	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(cc), config.WithRegion("us-west-2"))
	if err != nil {
		t.Fatal(err)
	}

	sm := secretsmanager.NewFromConfig(cfg, func(o *secretsmanager.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
	asm := ASM{Client: *sm}
	return asm
}

func TestASMWorkflow(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	asm := localstack(ctx, t)
	ref := Ref{Module: Some("foo"), Name: "bar"}
	var mySecret = []byte("my secret")
	manager, err := New(ctx, asm, []Provider[Secrets]{asm})
	assert.NoError(t, err)

	var gotSecret []byte
	err = manager.Get(ctx, ref, &gotSecret)
	assert.Error(t, err)

	items, err := manager.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, items, []Entry{})

	err = manager.Set(ctx, "asm", ref, mySecret)
	assert.NoError(t, err)

	items, err = manager.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, items, []Entry{{Ref: ref, Accessor: URL("asm://foo.bar")}})

	err = manager.Get(ctx, ref, &gotSecret)
	assert.NoError(t, err)

	// Set again to make sure it updates.
	mySecret = []byte("hunter1")
	err = manager.Set(ctx, "asm", ref, mySecret)
	assert.NoError(t, err)

	err = manager.Get(ctx, ref, &gotSecret)
	assert.NoError(t, err)
	assert.Equal(t, gotSecret, mySecret)

	err = manager.Unset(ctx, "asm", ref)
	assert.NoError(t, err)

	items, err = manager.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, items, []Entry{})

	err = manager.Get(ctx, ref, &gotSecret)
	assert.Error(t, err)
}

// Suggest not running this against a real AWS account (especially in CI) due to the cost. Maybe costs a few $.
func TestASMPagination(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	asm := localstack(ctx, t)
	manager, err := New(ctx, asm, []Provider[Secrets]{asm})
	assert.NoError(t, err)

	// Create 210 secrets, so we paginate at least twice.
	for i := range 210 {
		ref := NewRef("foo", fmt.Sprintf("bar%03d", i))
		err := manager.Set(ctx, "asm", ref, []byte(fmt.Sprintf("hunter%03d", i)))
		assert.NoError(t, err)
	}

	items, err := manager.List(ctx)
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
		var secret []byte
		err := manager.Get(ctx, item.Ref, &secret)
		assert.NoError(t, err)
		assert.Equal(t, secret, []byte(fmt.Sprintf("hunter%03d", i)))
	}

	// Delete them
	for i := range 210 {
		ref := NewRef("foo", fmt.Sprintf("bar%03d", i))
		err := manager.Unset(ctx, "asm", ref)
		assert.NoError(t, err)
	}

	// Make sure they are all gone
	items, err = manager.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(items), 0)
}
