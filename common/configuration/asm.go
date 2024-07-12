package configuration

import (
	"context"
	"net/url"
	"time"

	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/alecthomas/types/optional"
	"github.com/puzpuzpuz/xsync/v3"

	"github.com/TBD54566975/ftl/backend/controller/leader"
	"github.com/TBD54566975/ftl/backend/controller/leases"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/log"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type asmClient interface {
	syncInterval() time.Duration
	sync(ctx context.Context, values *xsync.MapOf[Ref, SyncedValue]) error
	store(ctx context.Context, ref Ref, value []byte) (*url.URL, error)
	delete(ctx context.Context, ref Ref) error
}

// ASM implements a Provider for AWS Secrets Manager (ASM).
// Only supports loading "string" secrets, not binary secrets.
//
// One controller is elected as the leader and is responsible for syncing the cache of secrets from ASM (see asmLeader).
// Others get secrets from the leader via AdminService (see asmFollower).
type ASM struct {
	coordinator *leader.Coordinator[asmClient]
}

var _ AsynchronousProvider[Secrets] = &ASM{}

func NewASM(ctx context.Context, secretsClient *secretsmanager.Client, advertise *url.URL, leaser leases.Leaser) *ASM {
	return newASMForTesting(ctx, secretsClient, advertise, leaser, optional.None[asmClient]())
}

func newASMForTesting(ctx context.Context, secretsClient *secretsmanager.Client, advertise *url.URL, leaser leases.Leaser, override optional.Option[asmClient]) *ASM {
	leaderFactory := func(ctx context.Context) (asmClient, error) {
		if override, ok := override.Get(); ok {
			return override, nil
		}
		return newASMLeader(ctx, secretsClient), nil
	}
	followerFactory := func(ctx context.Context, url *url.URL) (client asmClient, err error) {
		if override, ok := override.Get(); ok {
			return override, nil
		}
		rpcClient := rpc.Dial(ftlv1connect.NewAdminServiceClient, url.String(), log.Error)
		return newASMFollower(ctx, rpcClient), nil
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

func (a *ASM) SyncInterval() time.Duration {
	client, err := a.coordinator.Get()
	if err != nil {
		// Could not coordinate, try again soon
		return time.Second * 5
	}
	return client.syncInterval()
}

func (a *ASM) Sync(ctx context.Context, values *xsync.MapOf[Ref, SyncedValue]) error {
	client, err := a.coordinator.Get()
	if err != nil {
		return err
	}
	return client.sync(ctx, values)
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
