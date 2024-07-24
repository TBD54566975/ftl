//go:build integration

package configuration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"sort"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/TBD54566975/ftl/backend/controller/leases"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/slices"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
	. "github.com/alecthomas/types/optional"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

func setUp(ctx context.Context, t *testing.T, router optional.Option[Router[Secrets]]) (*Manager[Secrets], *ASM, *asmLeader, *secretsmanager.Client, *ManualSyncProvider[Secrets], *leases.FakeLeaser) {
	t.Helper()

	if _, ok := router.Get(); !ok {
		dir := t.TempDir()
		projectPath := path.Join(dir, "ftl-project.toml")
		os.WriteFile(projectPath, []byte(`name = "asmtest"`), 0600)
		router = optional.Some[Router[Secrets]](ProjectConfigResolver[Secrets]{Config: projectPath})
	}

	cc := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test", "test", ""))
	cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(cc), config.WithRegion("us-west-2"))
	assert.NoError(t, err)

	externalClient := secretsmanager.NewFromConfig(cfg, func(o *secretsmanager.Options) {
		o.BaseEndpoint = aws.String("http://localhost:4566")
	})
	leaser := leases.NewFakeLeaser()
	asm := NewASM(ctx, externalClient, URL("http://localhost:1234"), leaser)
	manualSyncProvider := NewManualSyncProvider[Secrets](asm)

	sm, err := New(ctx, router.MustGet(), []Provider[Secrets]{manualSyncProvider})
	assert.NoError(t, err)

	leaderOrFollower, err := asm.coordinator.Get()
	assert.NoError(t, err)
	leader, ok := leaderOrFollower.(*asmLeader)
	assert.True(t, ok, "expected test to get an asm leader not a follower")
	return sm, asm, leader, externalClient, manualSyncProvider, leaser
}

func waitForUpdatesToProcess(c *cache[Secrets]) {
	c.topicWaitGroup.Wait()
}

func TestASMWorkflow(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	sr := NewDBSecretResolver(&mockDBSecretResolverDAL{})
	sm, _, _, _, _, _ := setUp(ctx, t, Some[Router[Secrets]](sr))
	ref := Ref{Module: Some("foo"), Name: "bar"}
	var mySecret = jsonBytes(t, "my secret")

	// wait for initial sync to complete
	err := sm.cache.providers["asm"].waitForInitialSync()
	assert.NoError(t, err)

	var gotSecret []byte
	err = sm.Get(ctx, ref, &gotSecret)
	assert.Error(t, err)

	items, err := sm.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, items, []Entry{})

	err = sm.Set(ctx, "asm", ref, mySecret)
	waitForUpdatesToProcess(sm.cache)
	assert.NoError(t, err)

	items, err = sm.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, items, []Entry{{Ref: ref, Accessor: URL("asm://foo.bar")}})

	err = sm.Get(ctx, ref, &gotSecret)
	assert.NoError(t, err)

	// Set again to make sure it updates.
	mySecret = jsonBytes(t, "hunter1")
	err = sm.Set(ctx, "asm", ref, mySecret)
	waitForUpdatesToProcess(sm.cache)
	assert.NoError(t, err)

	err = sm.Get(ctx, ref, &gotSecret)
	assert.NoError(t, err)
	assert.Equal(t, gotSecret, mySecret)

	err = sm.Unset(ctx, "asm", ref)
	waitForUpdatesToProcess(sm.cache)
	assert.NoError(t, err)

	items, err = sm.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, items, []Entry{})

	err = sm.Get(ctx, ref, &gotSecret)
	assert.Error(t, err)
}

// Suggest not running this against a real AWS account (especially in CI) due to the cost. Maybe costs a few $.
func TestASMPagination(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	sm, asm, _, _, _, _ := setUp(ctx, t, None[Router[Secrets]]())
	sr := NewDBSecretResolver(&mockDBSecretResolverDAL{})
	manager, err := New(ctx, sr, []Provider[Secrets]{asm})
	assert.NoError(t, err)

	// Create 210 secrets, so we paginate at least twice.
	for i := range 210 {
		ref := NewRef("foo", fmt.Sprintf("bar%03d", i))
		err := manager.Set(ctx, "asm", ref, jsonBytes(t, fmt.Sprintf("hunter%03d", i)))
		assert.NoError(t, err)
	}

	waitForUpdatesToProcess(sm.cache)

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
		assert.Equal(t, secret, jsonBytes(t, fmt.Sprintf("hunter%03d", i)))
	}

	// Delete them
	for i := range 210 {
		ref := NewRef("foo", fmt.Sprintf("bar%03d", i))
		err := manager.Unset(ctx, "asm", ref)
		assert.NoError(t, err)
	}
	waitForUpdatesToProcess(sm.cache)

	// Make sure they are all gone
	items, err = manager.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(items), 0)
}

// TestLeaderSync sets and gets values via the leader, as well as directly with ASM
func TestLeaderSync(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	sm, _, _, externalClient, manualSync, _ := setUp(ctx, t, None[Router[Secrets]]())
	testClientSync(ctx, t, sm, externalClient, true, []*ManualSyncProvider[Secrets]{manualSync})
}

// TestFollowerSync tests setting and getting values from a follower to the leader to ASM, and vice versa
func TestFollowerSync(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	leaderManager, _, _, externalClient, leaderManualSync, leaser := setUp(ctx, t, None[Router[Secrets]]())

	// fakeRPCClient connects the follower to the leader
	fakeRPCClient := &fakeAdminClient{sm: leaderManager}
	follower := newASMFollower(fakeRPCClient, "fake", time.Second)

	followerASM := newASMForTesting(ctx, externalClient, URL("http://localhost:1235"), leaser, optional.Some[asmClient](follower))
	asmClient, err := followerASM.coordinator.Get()
	assert.NoError(t, err)
	_, ok := asmClient.(*asmFollower)
	assert.True(t, ok, "expected test to get an asm follower not a leader")

	followerManualSync := NewManualSyncProvider(followerASM)

	sm, err := New(ctx, leaderManager.router, []Provider[Secrets]{followerManualSync})
	assert.NoError(t, err)

	testClientSync(ctx, t, sm, externalClient, false, []*ManualSyncProvider[Secrets]{leaderManualSync, followerManualSync})
}

// testClientSync uses a Manager and a secretsmanager.Client to test setting and getting secrets
func testClientSync(ctx context.Context,
	t *testing.T,
	sm *Manager[Secrets],
	externalClient *secretsmanager.Client,
	isLeader bool,
	manualSyncProviders []*ManualSyncProvider[Secrets]) {
	t.Helper()

	waitForManualSync(t, manualSyncProviders)

	// write a secret via asmClient
	clientRef := Ref{Module: Some("sync"), Name: "set-by-client"}
	err := sm.Set(ctx, "asm", clientRef, "client-first")
	assert.NoError(t, err)
	waitForUpdatesToProcess(sm.cache)
	value, err := sm.getData(ctx, clientRef)
	assert.NoError(t, err, "failed to load secret via asm")
	assert.Equal(t, value, jsonBytes(t, "client-first"), "unexpected secret value")

	// write another secret via sm directly
	smRef := Ref{Module: Some("sync"), Name: "set-by-sm"}
	err = storeUnobfuscatedValueInASM(ctx, sm, externalClient, smRef, []byte(jsonString(t, "sm-first")), true)
	assert.NoError(t, err, "failed to create secret via sm")
	waitForUpdatesToProcess(sm.cache)
	value, err = sm.getData(ctx, smRef)
	assert.Error(t, err, "expected to fail because asm client has not synced secret yet")

	// write a secret via client and then by sm directly
	clientSmRef := Ref{Module: Some("sync"), Name: "set-by-client-then-sm"}
	err = sm.Set(ctx, "asm", clientSmRef, "client-sm-first")
	assert.NoError(t, err)
	err = storeUnobfuscatedValueInASM(ctx, sm, externalClient, clientSmRef, []byte(jsonString(t, "client-sm-second")), false)
	assert.NoError(t, err)
	waitForUpdatesToProcess(sm.cache)
	value, err = sm.getData(ctx, clientSmRef)
	assert.NoError(t, err, "failed to load secret via asm")
	assert.Equal(t, value, jsonBytes(t, "client-sm-first"), "expected initial value before client has a chance to sync newest value")

	// write a secret via sm directly and then by client
	smClientRef := Ref{Module: Some("sync"), Name: "set-by-sm-then-client"}
	err = storeUnobfuscatedValueInASM(ctx, sm, externalClient, smClientRef, []byte(jsonString(t, "sm-client-first")), true)
	assert.NoError(t, err, "failed to create secret via sm")
	err = sm.Set(ctx, "asm", smClientRef, "sm-client-second")
	assert.NoError(t, err)
	waitForUpdatesToProcess(sm.cache)
	value, err = sm.getData(ctx, smClientRef)
	assert.NoError(t, err, "failed to load secret via asm")
	assert.Equal(t, value, jsonBytes(t, "sm-client-second"), "unexpected secret value")

	waitForManualSync(t, manualSyncProviders)

	// confirm that all secrets are up to date
	list, err := sm.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, slices.Sort(slices.Map(list, func(e Entry) string { return e.Ref.String() })),
		[]string{"sync.set-by-client", "sync.set-by-client-then-sm", "sync.set-by-sm", "sync.set-by-sm-then-client"})
	for _, entry := range list {
		value, err = sm.getData(ctx, entry.Ref)
		assert.NoError(t, err, "failed to load secret via asm")
		var expectedValue string
		switch entry.Ref {
		case clientRef:
			expectedValue = jsonString(t, "client-first")
		case smRef:
			expectedValue = jsonString(t, "sm-first")
		case clientSmRef:
			expectedValue = jsonString(t, "client-sm-second")
		case smClientRef:
			expectedValue = jsonString(t, "sm-client-second")
		default:
			t.Fatal(fmt.Sprintf("unexpected ref: %s", entry.Ref))
		}
		assert.Equal(t, expectedValue, string(value), "unexpected secret value for %s", entry.Ref)
	}

	if !isLeader {
		// only test deleting secrets causing errors in ASM Leader.
		// when leader starts returning errors for list, follower will not get updates
		return
	}

	// delete 2 secrets without client knowing
	tr := true
	_, err = externalClient.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(smRef.String()),
		ForceDeleteWithoutRecovery: &tr,
	})
	assert.NoError(t, err)
	_, err = externalClient.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(smClientRef.String()),
		ForceDeleteWithoutRecovery: &tr,
	})
	assert.NoError(t, err)

	waitForManualSync(t, manualSyncProviders)

	_, err = sm.getData(ctx, smRef)
	assert.Error(t, err, "expected to fail because secret was deleted")
	_, err = sm.getData(ctx, smClientRef)
	assert.Error(t, err, "expected to fail because secret was deleted")
}

// waitForManualSync syncs each provider in order
func waitForManualSync[R Role](t *testing.T, providers []*ManualSyncProvider[R]) {
	t.Helper()

	for _, provider := range providers {
		err := provider.SyncAndWait()
		assert.NoError(t, err)
	}
}

func storeUnobfuscatedValueInASM(ctx context.Context, sm *Manager[Secrets], externalClient *secretsmanager.Client, ref Ref, value []byte, isNew bool) error {
	obfuscator := Secrets{}.obfuscator()
	obfuscatedValue, err := obfuscator.Obfuscate(value)
	if err != nil {
		return err
	}
	if isNew {
		_, err = externalClient.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
			Name:         aws.String(ref.String()),
			SecretString: aws.String(string(obfuscatedValue)),
			Tags: []types.Tag{
				{Key: aws.String(asmTagKey), Value: aws.String(ref.Module.Default(""))},
			},
		})
	} else {
		_, err = externalClient.UpdateSecret(ctx, &secretsmanager.UpdateSecretInput{
			SecretId:     aws.String(ref.String()),
			SecretString: aws.String(string(obfuscatedValue)),
		})
	}
	if err != nil {
		return err
	}

	// update router so ref is in list. This simulates another controller added the ref
	url, err := url.Parse("asm://set-directly-through-asm")
	if err != nil {
		return err
	}
	return sm.router.Set(ctx, ref, url)
}

// jsonBytes is a helper to take a string value and convert it into json bytes
func jsonBytes(t *testing.T, value string) []byte {
	t.Helper()
	json, err := json.Marshal(value)
	assert.NoError(t, err, "failed to marshal value")
	return []byte(string(json))
}

// jsonString is a helper to take a string value and convert it into a json string
func jsonString(t *testing.T, value string) string {
	t.Helper()
	return string(jsonBytes(t, value))
}

// fakeAdminClient is a fake implementation of the AdminClient interface to allow tests to connect an Manager with an ASM Follower to a Manager with an ASM Leader
type fakeAdminClient struct {
	sm *Manager[Secrets]
}

func (c *fakeAdminClient) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

// ConfigList returns the list of configuration values, optionally filtered by module.
func (c *fakeAdminClient) ConfigList(ctx context.Context, req *connect.Request[ftlv1.ListConfigRequest]) (*connect.Response[ftlv1.ListConfigResponse], error) {
	panic("not implemented")
}

func (c *fakeAdminClient) ConfigGet(ctx context.Context, req *connect.Request[ftlv1.GetConfigRequest]) (*connect.Response[ftlv1.GetConfigResponse], error) {
	panic("not implemented")
}

func (c *fakeAdminClient) ConfigSet(ctx context.Context, req *connect.Request[ftlv1.SetConfigRequest]) (*connect.Response[ftlv1.SetConfigResponse], error) {
	panic("not implemented")
}

func (c *fakeAdminClient) ConfigUnset(ctx context.Context, req *connect.Request[ftlv1.UnsetConfigRequest]) (*connect.Response[ftlv1.UnsetConfigResponse], error) {
	panic("not implemented")
}

func (c *fakeAdminClient) SecretsList(ctx context.Context, req *connect.Request[ftlv1.ListSecretsRequest]) (*connect.Response[ftlv1.ListSecretsResponse], error) {
	listing, err := c.sm.List(ctx)
	if err != nil {
		return nil, err
	}
	secrets := []*ftlv1.ListSecretsResponse_Secret{}
	for _, secret := range listing {
		module, ok := secret.Module.Get()
		if *req.Msg.Module != "" && module != *req.Msg.Module {
			continue
		}
		ref := secret.Name
		if ok {
			ref = fmt.Sprintf("%s.%s", module, secret.Name)
		}
		var sv []byte
		if *req.Msg.IncludeValues {
			sv, err = c.sm.getData(ctx, secret.Ref)
			if err != nil {
				return nil, err
			}
		}
		secrets = append(secrets, &ftlv1.ListSecretsResponse_Secret{
			RefPath: ref,
			Value:   sv,
		})
	}
	return connect.NewResponse(&ftlv1.ListSecretsResponse{Secrets: secrets}), nil
}

// SecretGet returns the secret value for a given ref string.
func (c *fakeAdminClient) SecretGet(ctx context.Context, req *connect.Request[ftlv1.GetSecretRequest]) (*connect.Response[ftlv1.GetSecretResponse], error) {
	ref := NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name)
	v, err := c.sm.getData(ctx, ref)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.GetSecretResponse{Value: v}), nil
}

// SecretSet sets the secret at the given ref to the provided value.
func (c *fakeAdminClient) SecretSet(ctx context.Context, req *connect.Request[ftlv1.SetSecretRequest]) (*connect.Response[ftlv1.SetSecretResponse], error) {
	err := c.sm.SetJSON(ctx, "asm", NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name), req.Msg.Value)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.SetSecretResponse{}), nil
}

// SecretUnset unsets the secret value at the given ref.
func (c *fakeAdminClient) SecretUnset(ctx context.Context, req *connect.Request[ftlv1.UnsetSecretRequest]) (*connect.Response[ftlv1.UnsetSecretResponse], error) {
	err := c.sm.Unset(ctx, "asm", NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.UnsetSecretResponse{}), nil
}
