package configuration

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/benbjohnson/clock"

	"github.com/TBD54566975/ftl/backend/controller/leader"
	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/log"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type asmClient interface {
	list(ctx context.Context) ([]Entry, error)
	load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error)
	store(ctx context.Context, ref Ref, value []byte) (*url.URL, error)
	delete(ctx context.Context, ref Ref) error
}

// ASM implements Router and Provider for AWS Secrets Manager (ASM).
// Only supports loading "string" secrets, not binary secrets.
//
// The router does a direct/proxy map from a Ref to a URL, module.name <-> asm://module.name and does not access ASM at all.
//
// One controller is elected as the leader and is responsible for syncing the cache of secrets from ASM (see asmLeader).
// Others get secrets from the leader via AdminService (see asmFollower).
type ASM struct {
	coordinator *leader.Coordinator[asmClient]
}

var _ Router[Secrets] = &ASM{}
var _ Provider[Secrets] = &ASM{}

func NewASM(ctx context.Context, secretsClient *secretsmanager.Client, advertise *url.URL, leaser leases.Leaser) *ASM {
	return newASMForTesting(ctx, secretsClient, advertise, leaser, clock.New())
}

func newASMForTesting(ctx context.Context, secretsClient *secretsmanager.Client, advertise *url.URL, leaser leases.Leaser, clock clock.Clock) *ASM {
	leaderFactory := func(ctx context.Context) (asmClient, error) {
		return newASMLeader(ctx, secretsClient, clock), nil
	}
	followerFactory := func(ctx context.Context, url *url.URL) (client asmClient, err error) {
		rpcClient := rpc.Dial(ftlv1connect.NewAdminServiceClient, url.String(), log.Error)
		return newASMFollower(ctx, rpcClient, clock), nil
	}
	return &ASM{
		coordinator: leader.NewCoordinator[asmClient](
			ctx,
			advertise,
			leases.SystemKey("asm"),
			leaser,
			time.Second*10,
			leaderFactory,
			followerFactory,
		),
	}
}

func asmURLForRef(ref Ref) *url.URL {
	return &url.URL{
		Scheme: "asm",
		Host:   ref.String(),
	}
}

func (ASM) Role() Secrets {
	return Secrets{}
}

func (ASM) Key() string {
	return "asm"
}

func (ASM) Get(ctx context.Context, ref Ref) (*url.URL, error) {
	return asmURLForRef(ref), nil
}

func (ASM) Set(ctx context.Context, ref Ref, key *url.URL) error {
	expectedKey := asmURLForRef(ref)
	if key.String() != expectedKey.String() {
		return fmt.Errorf("key does not match expected key for ref: %s", expectedKey)
	}

	return nil
}

func (ASM) Unset(ctx context.Context, ref Ref) error {
	// removing a secret is handled in Delete()
	return nil
}

// List all secrets in the account
func (a *ASM) List(ctx context.Context) ([]Entry, error) {
	client, err := a.coordinator.Get()
	if err != nil {
		return nil, err
	}
	return client.list(ctx)
}

func (a *ASM) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	client, err := a.coordinator.Get()
	if err != nil {
		return nil, err
	}
	return client.load(ctx, ref, key)
}

// Store and if the secret already exists, update it.
func (a *ASM) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	client, err := a.coordinator.Get()
	if err != nil {
		return nil, err
	}
	return client.store(ctx, ref, value)
}

func (a *ASM) Delete(ctx context.Context, ref Ref) error {
	client, err := a.coordinator.Get()
	if err != nil {
		return err
	}
	return client.delete(ctx, ref)
}

func (a *ASM) UseWithProvider(ctx context.Context, pkey string) bool {
	return pkey == a.Key()
}
