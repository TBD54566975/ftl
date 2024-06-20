//go:build integration

package configuration

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/benbjohnson/clock"

	"github.com/alecthomas/assert/v2"
	. "github.com/alecthomas/types/optional"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func localstack(ctx context.Context, t *testing.T) (*ASM, *asmLeader, *secretsmanager.Client, *clock.Mock) {
	t.Helper()
	mockClock := clock.NewMock()
	cc := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test", "test", ""))
	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(cc), config.WithRegion("us-west-2"))
	if err != nil {
		t.Fatal(err)
	}
	sm := secretsmanager.NewFromConfig(cfg, func(o *secretsmanager.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
	asm := newASMForTesting(ctx, sm, URL("http://localhost:1234"), leases.NewFakeLeaser(), mockClock)

	leaderOrFollower, err := asm.coordinator.Get()
	assert.NoError(t, err)
	leader, ok := leaderOrFollower.(*asmLeader)
	assert.True(t, ok, "expected test to get an asm leader not a follower")
	return asm, leader, sm, mockClock
}

func waitForUpdatesToProcess(l *asmLeader) {
	l.topicWaitGroup.Wait()
}

func TestASMWorkflow(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	asm, leader, _, _ := localstack(ctx, t)
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
	waitForUpdatesToProcess(leader)
	assert.NoError(t, err)

	items, err = manager.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, items, []Entry{{Ref: ref, Accessor: URL("asm://foo.bar")}})

	err = manager.Get(ctx, ref, &gotSecret)
	assert.NoError(t, err)

	// Set again to make sure it updates.
	mySecret = []byte("hunter1")
	err = manager.Set(ctx, "asm", ref, mySecret)
	waitForUpdatesToProcess(leader)
	assert.NoError(t, err)

	err = manager.Get(ctx, ref, &gotSecret)
	assert.NoError(t, err)
	assert.Equal(t, gotSecret, mySecret)

	err = manager.Unset(ctx, "asm", ref)
	waitForUpdatesToProcess(leader)
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
	asm, leader, _, _ := localstack(ctx, t)
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
	waitForUpdatesToProcess(leader)

	// Make sure they are all gone
	items, err = manager.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(items), 0)
}

func TestSync(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	asm, leader, sm, clock := localstack(ctx, t)

	// wait for initial load
	err := leader.waitForSecrets()
	assert.NoError(t, err)

	// advance clock to half way between syncs
	clock.Add(syncInterval / 2)

	// write a secret via leader
	leaderRef := Ref{Module: Some("sync"), Name: "set-by-leader"}
	_, err = asm.Store(ctx, leaderRef, []byte("leader-first"))
	assert.NoError(t, err)
	waitForUpdatesToProcess(leader)
	value, err := asm.Load(ctx, leaderRef, asmURLForRef(leaderRef))
	assert.NoError(t, err, "failed to load secret via asm")
	assert.Equal(t, value, []byte("leader-first"), "unexpected secret value")

	// write another secret via sm directly
	smRef := Ref{Module: Some("sync"), Name: "set-by-sm"}
	_, err = sm.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(smRef.String()),
		SecretString: aws.String(string("sm-first")),
	})
	assert.NoError(t, err, "failed to create secret via sm")
	waitForUpdatesToProcess(leader)
	value, err = asm.Load(ctx, smRef, asmURLForRef(smRef))
	assert.Error(t, err, "expected to fail because asm leader has not synced secret yet")

	// write a secret via leader and then by sm directly
	leaderSmRef := Ref{Module: Some("sync"), Name: "set-by-leader-then-sm"}
	_, err = asm.Store(ctx, leaderSmRef, []byte("leader-sm-first"))
	assert.NoError(t, err)
	_, err = sm.UpdateSecret(ctx, &secretsmanager.UpdateSecretInput{
		SecretId:     aws.String(leaderSmRef.String()),
		SecretString: aws.String("leader-sm-second"),
	})
	assert.NoError(t, err)
	waitForUpdatesToProcess(leader)
	value, err = asm.Load(ctx, leaderSmRef, asmURLForRef(leaderSmRef))
	assert.NoError(t, err, "failed to load secret via asm")
	assert.Equal(t, value, []byte("leader-sm-first"), "expected initial value before leader has a chance to sync newest value")

	// write a secret via sm directly and then by leader
	smLeaderRef := Ref{Module: Some("sync"), Name: "set-by-sm-then-leader"}
	_, err = sm.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(smLeaderRef.String()),
		SecretString: aws.String(string("sm-leader-first")),
	})
	assert.NoError(t, err, "failed to create secret via sm")
	_, err = asm.Store(ctx, smLeaderRef, []byte("sm-leader-second"))
	assert.NoError(t, err)
	waitForUpdatesToProcess(leader)
	value, err = asm.Load(ctx, smLeaderRef, asmURLForRef(smLeaderRef))
	assert.NoError(t, err, "failed to load secret via asm")
	assert.Equal(t, value, []byte("sm-leader-second"), "unexpected secret value")

	// give leader a change to sync
	clock.Add(syncInterval)
	time.Sleep(time.Second * 5)

	// confirm that all secrets are up to date
	list, err := asm.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(list), 4, "expected 4 secrets")
	for _, entry := range list {
		value, err = asm.Load(ctx, entry.Ref, asmURLForRef(entry.Ref))
		assert.NoError(t, err, "failed to load secret via asm")
		var expectedValue string
		switch entry.Ref {
		case leaderRef:
			expectedValue = "leader-first"
		case smRef:
			expectedValue = "sm-first"
		case leaderSmRef:
			expectedValue = "leader-sm-second"
		case smLeaderRef:
			expectedValue = "sm-leader-second"
		default:
			t.Fatal(fmt.Sprintf("unexpected ref: %s", entry.Ref))
		}
		assert.Equal(t, expectedValue, string(value), "unexpected secret value for %s", entry.Ref)
	}

	// delete 2 secrets without leader knowing
	tr := true
	_, err = sm.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(smRef.String()),
		ForceDeleteWithoutRecovery: &tr,
	})
	assert.NoError(t, err)
	_, err = sm.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(smLeaderRef.String()),
		ForceDeleteWithoutRecovery: &tr,
	})
	assert.NoError(t, err)

	// give leader a change to sync
	clock.Add(syncInterval)
	time.Sleep(time.Second * 5)

	// confirm secrets were deleted
	list, err = asm.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(list), 2, "expected 2 secrets")
	_, err = asm.Load(ctx, smRef, asmURLForRef(smRef))
	assert.Error(t, err, "expected to fail because secret was deleted")
	_, err = asm.Load(ctx, smLeaderRef, asmURLForRef(smLeaderRef))
	assert.Error(t, err, "expected to fail because secret was deleted")
}
