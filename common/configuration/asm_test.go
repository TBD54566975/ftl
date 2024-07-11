//go:build integration

package configuration

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/TBD54566975/ftl/backend/controller/leases"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/benbjohnson/clock"

	"github.com/alecthomas/assert/v2"
	. "github.com/alecthomas/types/optional"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
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

func waitForUpdatesToProcess(c *secretsCache) {
	c.topicWaitGroup.Wait()
}

func TestASMWorkflow(t *testing.T) {
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	asm, leader, _, _ := localstack(ctx, t)
	ref := Ref{Module: Some("foo"), Name: "bar"}
	var mySecret = jsonBytes(t, "my secret")
	sr := NewDBSecretResolver(&mockDBSecretResolverDAL{})
	manager, err := New(ctx, sr, []Provider[Secrets]{asm})
	assert.NoError(t, err)

	var gotSecret []byte
	err = manager.Get(ctx, ref, &gotSecret)
	assert.Error(t, err)

	items, err := manager.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, items, []Entry{})

	err = manager.Set(ctx, "asm", ref, mySecret)
	waitForUpdatesToProcess(leader.cache)
	assert.NoError(t, err)

	items, err = manager.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, items, []Entry{{Ref: ref, Accessor: URL("asm://foo.bar")}})

	err = manager.Get(ctx, ref, &gotSecret)
	assert.NoError(t, err)

	// Set again to make sure it updates.
	mySecret = jsonBytes(t, "hunter1")
	err = manager.Set(ctx, "asm", ref, mySecret)
	waitForUpdatesToProcess(leader.cache)
	assert.NoError(t, err)

	err = manager.Get(ctx, ref, &gotSecret)
	assert.NoError(t, err)
	assert.Equal(t, gotSecret, mySecret)

	err = manager.Unset(ctx, "asm", ref)
	waitForUpdatesToProcess(leader.cache)
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
	sr := NewDBSecretResolver(&mockDBSecretResolverDAL{})
	manager, err := New(ctx, sr, []Provider[Secrets]{asm})
	assert.NoError(t, err)

	// Create 210 secrets, so we paginate at least twice.
	for i := range 210 {
		ref := NewRef("foo", fmt.Sprintf("bar%03d", i))
		err := manager.Set(ctx, "asm", ref, jsonBytes(t, fmt.Sprintf("hunter%03d", i)))
		assert.NoError(t, err)
	}

	waitForUpdatesToProcess(leader.cache)

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
	waitForUpdatesToProcess(leader.cache)

	// Make sure they are all gone
	items, err = manager.List(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(items), 0)
}

func TestLeaderSync(t *testing.T) {
	// test setting and getting values via the leader, as well as directly with ASM
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	_, leader, sm, clock := localstack(ctx, t)
	testClientSync(ctx, t, leader, leader.cache, sm, func(percentage float64) {
		clock.Add(time.Duration(percentage) * asmLeaderSyncInterval)
	})
}

func TestFollowerSync(t *testing.T) {
	// Test setting and getting values via the follower, which is connected to a leader, as well as directly with ASM
	ctx := log.ContextWithNewDefaultLogger(context.Background())
	asm, _, sm, leaderClock := localstack(ctx, t)

	// fakeRPCClient connects the follower to the leader
	fakeRPCClient := &fakeAdminClient{asm: asm}
	followerClock := clock.NewMock()
	follower := newASMFollower(ctx, fakeRPCClient, "fake", followerClock)

	testClientSync(ctx, t, follower, follower.cache, sm, func(percentage float64) {
		// sync leader
		leaderClock.Add(time.Duration(percentage) * asmLeaderSyncInterval)
		if percentage == 1.0 {
			time.Sleep(time.Second * 5)
		}

		// then sync follower
		followerClock.Add(time.Duration(percentage) * asmFollowerSyncInterval)
	})
}

func testClientSync(ctx context.Context,
	t *testing.T,
	client asmClient,
	cache *secretsCache,
	externalClient *secretsmanager.Client,
	progressByIntervalPercentage func(percentage float64)) {
	t.Helper()

	// wait for initial load
	err := cache.waitForSecrets()
	assert.NoError(t, err)

	// advance clock to half way between syncs
	progressByIntervalPercentage(0.5)

	// write a secret via asmClient
	clientRef := Ref{Module: Some("sync"), Name: "set-by-client"}
	err = storeUnobfuscatedValue(ctx, client, clientRef, jsonBytes(t, "client-first"))
	assert.NoError(t, err)
	waitForUpdatesToProcess(cache)
	value, err := getUnobfuscatedValue(ctx, client, clientRef)
	assert.NoError(t, err, "failed to load secret via asm")
	assert.Equal(t, value, jsonBytes(t, "client-first"), "unexpected secret value")

	// write another secret via sm directly
	smRef := Ref{Module: Some("sync"), Name: "set-by-sm"}
	err = storeUnobfuscatedValueInASM(ctx, externalClient, smRef, []byte(jsonString(t, "sm-first")), true)
	assert.NoError(t, err, "failed to create secret via sm")
	waitForUpdatesToProcess(cache)
	value, err = getUnobfuscatedValue(ctx, client, smRef)
	assert.Error(t, err, "expected to fail because asm client has not synced secret yet")

	// write a secret via client and then by sm directly
	clientSmRef := Ref{Module: Some("sync"), Name: "set-by-client-then-sm"}
	err = storeUnobfuscatedValue(ctx, client, clientSmRef, jsonBytes(t, "client-sm-first"))
	assert.NoError(t, err)
	err = storeUnobfuscatedValueInASM(ctx, externalClient, clientSmRef, []byte(jsonString(t, "client-sm-second")), false)
	assert.NoError(t, err)
	waitForUpdatesToProcess(cache)
	value, err = getUnobfuscatedValue(ctx, client, clientSmRef)
	assert.NoError(t, err, "failed to load secret via asm")
	assert.Equal(t, value, jsonBytes(t, "client-sm-first"), "expected initial value before client has a chance to sync newest value")

	// write a secret via sm directly and then by client
	smClientRef := Ref{Module: Some("sync"), Name: "set-by-sm-then-client"}
	err = storeUnobfuscatedValueInASM(ctx, externalClient, smClientRef, []byte(jsonString(t, "sm-client-first")), true)
	assert.NoError(t, err, "failed to create secret via sm")
	err = storeUnobfuscatedValue(ctx, client, smClientRef, jsonBytes(t, "sm-client-second"))
	assert.NoError(t, err)
	waitForUpdatesToProcess(cache)
	value, err = getUnobfuscatedValue(ctx, client, smClientRef)
	assert.NoError(t, err, "failed to load secret via asm")
	assert.Equal(t, value, jsonBytes(t, "sm-client-second"), "unexpected secret value")

	// give client a change to sync
	progressByIntervalPercentage(1.0)
	time.Sleep(time.Second * 5)

	// confirm that all secrets are up to date
	list, err := client.list(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(list), 4, "expected 4 secrets")
	for _, entry := range list {
		value, err = getUnobfuscatedValue(ctx, client, entry.Ref)
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

	// give client a change to sync
	progressByIntervalPercentage(1.0)
	time.Sleep(time.Second * 5)

	// confirm secrets were deleted
	list, err = client.list(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(list), 2, "expected 2 secrets")
	_, err = getUnobfuscatedValue(ctx, client, smRef)
	assert.Error(t, err, "expected to fail because secret was deleted")
	_, err = getUnobfuscatedValue(ctx, client, smClientRef)
	assert.Error(t, err, "expected to fail because secret was deleted")
}

func storeUnobfuscatedValue(ctx context.Context, client asmClient, ref Ref, value []byte) error {
	obfuscator := Secrets{}.obfuscator()
	obfuscatedValue, err := obfuscator.Obfuscate(value)
	if err != nil {
		return err
	}
	_, err = client.store(ctx, ref, obfuscatedValue)
	return err
}

func getUnobfuscatedValue(ctx context.Context, client asmClient, ref Ref) ([]byte, error) {
	obfuscator := Secrets{}.obfuscator()
	obfuscatedValue, err := client.load(ctx, ref, asmURLForRef(ref))
	if err != nil {
		return nil, err
	}
	unobfuscatedValue, err := obfuscator.Reveal(obfuscatedValue)
	if err != nil {
		return nil, err
	}
	return unobfuscatedValue, nil
}

func storeUnobfuscatedValueInASM(ctx context.Context, externalClient *secretsmanager.Client, ref Ref, value []byte, isNew bool) error {
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
	return err
}

func jsonBytes(t *testing.T, value string) []byte {
	t.Helper()
	json, err := json.Marshal(value)
	assert.NoError(t, err, "failed to marshal value")
	return []byte(string(json))
}

func jsonString(t *testing.T, value string) string {
	t.Helper()
	return string(jsonBytes(t, value))
}

type fakeAdminClient struct {
	asm *ASM
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
	obfuscator := Secrets{}.obfuscator()
	client, err := c.asm.coordinator.Get()
	if err != nil {
		return nil, err
	}
	listing, err := client.list(ctx)
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
			obfuscatedValue, err := c.asm.Load(ctx, secret.Ref, asmURLForRef(secret.Ref))
			if err != nil {
				return nil, err
			}
			sv, err = obfuscator.Reveal(obfuscatedValue)
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
	obfuscator := Secrets{}.obfuscator()
	ref := NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name)
	obfuscatedValue, err := c.asm.Load(ctx, ref, asmURLForRef(ref))
	if err != nil {
		return nil, err
	}
	vb, err := obfuscator.Reveal(obfuscatedValue)
	return connect.NewResponse(&ftlv1.GetSecretResponse{Value: vb}), nil
}

// SecretSet sets the secret at the given ref to the provided value.
func (c *fakeAdminClient) SecretSet(ctx context.Context, req *connect.Request[ftlv1.SetSecretRequest]) (*connect.Response[ftlv1.SetSecretResponse], error) {
	obfuscator := Secrets{}.obfuscator()
	obfuscatedValue, err := obfuscator.Obfuscate(req.Msg.Value)
	if err != nil {
		return nil, err
	}
	_, err = c.asm.Store(ctx, NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name), obfuscatedValue)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.SetSecretResponse{}), nil
}

// SecretUnset unsets the secret value at the given ref.
func (c *fakeAdminClient) SecretUnset(ctx context.Context, req *connect.Request[ftlv1.UnsetSecretRequest]) (*connect.Response[ftlv1.UnsetSecretResponse], error) {
	err := c.asm.Delete(ctx, NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.UnsetSecretResponse{}), nil
}
