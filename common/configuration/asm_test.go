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
	"github.com/TBD54566975/ftl/common/configuration/sql"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/benbjohnson/clock"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
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
	sr := NewDBSecretResolver(&fakeDBSecretResolverDAL{})
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
	sr := NewDBSecretResolver(&fakeDBSecretResolverDAL{})
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
	sr := NewDBSecretResolver(&fakeDBSecretResolverDAL{})

	// fakeRPCClient connects the follower to the leader
	fakeRPCClient := &fakeAdminClient{asm: asm, sr: &sr}
	followerClock := clock.NewMock()
	follower := newASMFollower(ctx, fakeRPCClient, followerClock)

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
	sm *secretsmanager.Client,
	progressByIntervalPercentage func(percentage float64)) {
	t.Helper()

	// wait for initial load
	err := cache.waitForSecrets()
	assert.NoError(t, err)

	// advance clock to half way between syncs
	progressByIntervalPercentage(0.5)

	// write a secret via asmClient
	clientRef := Ref{Module: Some("sync"), Name: "set-by-client"}
	_, err = client.store(ctx, clientRef, jsonBytes(t, "client-first"))
	assert.NoError(t, err)
	waitForUpdatesToProcess(cache)
	value, err := client.load(ctx, clientRef, asmURLForRef(clientRef))
	assert.NoError(t, err, "failed to load secret via asm")
	assert.Equal(t, value, jsonBytes(t, "client-first"), "unexpected secret value")

	// write another secret via sm directly
	smRef := Ref{Module: Some("sync"), Name: "set-by-sm"}
	_, err = sm.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(smRef.String()),
		SecretString: aws.String(jsonString(t, "sm-first")),
		Tags: []types.Tag{
			{Key: aws.String(asmTagKey), Value: aws.String(smRef.Module.Default(""))},
		},
	})
	assert.NoError(t, err, "failed to create secret via sm")
	waitForUpdatesToProcess(cache)
	value, err = client.load(ctx, smRef, asmURLForRef(smRef))
	assert.Error(t, err, "expected to fail because asm client has not synced secret yet")

	// write a secret via client and then by sm directly
	clientSmRef := Ref{Module: Some("sync"), Name: "set-by-client-then-sm"}
	_, err = client.store(ctx, clientSmRef, jsonBytes(t, "client-sm-first"))
	assert.NoError(t, err)
	_, err = sm.UpdateSecret(ctx, &secretsmanager.UpdateSecretInput{
		SecretId:     aws.String(clientSmRef.String()),
		SecretString: aws.String(jsonString(t, "client-sm-second")),
	})
	assert.NoError(t, err)
	waitForUpdatesToProcess(cache)
	value, err = client.load(ctx, clientSmRef, asmURLForRef(clientSmRef))
	assert.NoError(t, err, "failed to load secret via asm")
	assert.Equal(t, value, jsonBytes(t, "client-sm-first"), "expected initial value before client has a chance to sync newest value")

	// write a secret via sm directly and then by client
	smClientRef := Ref{Module: Some("sync"), Name: "set-by-sm-then-client"}
	_, err = sm.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
		Name:         aws.String(smClientRef.String()),
		SecretString: aws.String(jsonString(t, "sm-client-first")),
		Tags: []types.Tag{
			{Key: aws.String(asmTagKey), Value: aws.String(smClientRef.Module.Default(""))},
		},
	})
	assert.NoError(t, err, "failed to create secret via sm")
	_, err = client.store(ctx, smClientRef, jsonBytes(t, "sm-client-second"))
	assert.NoError(t, err)
	waitForUpdatesToProcess(cache)
	value, err = client.load(ctx, smClientRef, asmURLForRef(smClientRef))
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
		value, err = client.load(ctx, entry.Ref, asmURLForRef(entry.Ref))
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
	_, err = sm.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(smRef.String()),
		ForceDeleteWithoutRecovery: &tr,
	})
	assert.NoError(t, err)
	_, err = sm.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
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
	_, err = client.load(ctx, smRef, asmURLForRef(smRef))
	assert.Error(t, err, "expected to fail because secret was deleted")
	_, err = client.load(ctx, smClientRef, asmURLForRef(smClientRef))
	assert.Error(t, err, "expected to fail because secret was deleted")
}

func jsonBytes(t *testing.T, value string) []byte {
	t.Helper()
	json, err := json.Marshal(value)
	assert.NoError(t, err, "failed to marshal value")
	return []byte("c" + string(json))
}

func jsonString(t *testing.T, value string) string {
	t.Helper()
	return string(jsonBytes(t, value))
}

type fakeAdminClient struct {
	asm *ASM
	sr  *DBSecretResolver
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
	listing, err := c.sr.List(ctx)
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
			sv, err = c.asm.Load(ctx, secret.Ref, asmURLForRef(secret.Ref))
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
	vb, err := c.asm.Load(ctx, ref, asmURLForRef(ref))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.GetSecretResponse{Value: vb}), nil
}

// SecretSet sets the secret at the given ref to the provided value.
func (c *fakeAdminClient) SecretSet(ctx context.Context, req *connect.Request[ftlv1.SetSecretRequest]) (*connect.Response[ftlv1.SetSecretResponse], error) {
	_, err := c.asm.Store(ctx, NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name), req.Msg.Value)
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

type fakeDBSecretResolverDAL struct {
	entries []sql.ModuleSecret
}

func (d *fakeDBSecretResolverDAL) findEntry(module Option[string], name string) (Option[sql.ModuleSecret], int) {
	for i := range d.entries {
		if d.entries[i].Module.Default("") == module.Default("") && d.entries[i].Name == name {
			return optional.Some(d.entries[i]), i
		}
	}
	return optional.None[sql.ModuleSecret](), -1
}

func (d *fakeDBSecretResolverDAL) GetModuleSecretURL(ctx context.Context, module Option[string], name string) (string, error) {
	entry, _ := d.findEntry(module, name)
	if e, ok := entry.Get(); ok {
		return e.Url, nil
	}
	return "", fmt.Errorf("secret not found")
}

func (d *fakeDBSecretResolverDAL) ListModuleSecrets(ctx context.Context) ([]sql.ModuleSecret, error) {
	return d.entries, nil
}

func (d *fakeDBSecretResolverDAL) SetModuleSecretURL(ctx context.Context, module Option[string], name string, url string) error {
	d.UnsetModuleSecret(ctx, module, name)
	d.entries = append(d.entries, sql.ModuleSecret{Module: module, Name: name, Url: url})
	return nil
}

func (d *fakeDBSecretResolverDAL) UnsetModuleSecret(ctx context.Context, module Option[string], name string) error {
	entry, i := d.findEntry(module, name)
	if _, ok := entry.Get(); ok {
		d.entries = append(d.entries[:i], d.entries[i+1:]...)
	}
	return nil
}
