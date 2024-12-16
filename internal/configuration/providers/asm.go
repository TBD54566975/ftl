package providers

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/alecthomas/types/optional"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/block/ftl/backend/controller/leases"
	"github.com/block/ftl/internal/configuration"
)

const ASMProviderKey configuration.ProviderKey = "asm"

type asmClient interface {
	name() string
	sync(ctx context.Context) (map[configuration.Ref]configuration.SyncedValue, error)
	store(ctx context.Context, ref configuration.Ref, value []byte) (*url.URL, error)
	delete(ctx context.Context, ref configuration.Ref) error
}

// ASM implements a Provider for AWS Secrets Manager (ASM).
// Only supports loading "string" secrets, not binary secrets.
//
// One controller is elected as the leader and is responsible for syncing the cache of secrets from ASM (see asmManager).
// Others get secrets from the leader via AdminService (see asmFollower).
type ASM struct {
	client asmClient
}

var _ configuration.AsynchronousProvider[configuration.Secrets] = &ASM{}

func NewASMFactory(secretsClient *secretsmanager.Client, advertise *url.URL, leaser leases.Leaser) (configuration.ProviderKey, Factory[configuration.Secrets]) {
	return ASMProviderKey, func(ctx context.Context) (configuration.Provider[configuration.Secrets], error) {
		return NewASM(secretsClient), nil
	}
}

func NewASM(secretsClient *secretsmanager.Client) *ASM {
	return newASMForTesting(secretsClient, optional.None[asmClient]())
}

func newASMForTesting(secretsClient *secretsmanager.Client, override optional.Option[asmClient]) *ASM {
	if override, ok := override.Get(); ok {
		return &ASM{
			client: override,
		}
	}
	return &ASM{
		client: newAsmManager(secretsClient),
	}
}

func asmURLForRef(ref configuration.Ref) *url.URL {
	return &url.URL{
		Scheme: string(ASMProviderKey),
		Host:   ref.String(),
	}
}

func (ASM) Role() configuration.Secrets {
	return configuration.Secrets{}
}

func (*ASM) Key() configuration.ProviderKey {
	return ASMProviderKey
}

func (a *ASM) SyncInterval() time.Duration {
	return time.Second * 5
}

func (a *ASM) Sync(ctx context.Context) (map[configuration.Ref]configuration.SyncedValue, error) {
	values, err := a.client.sync(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", a.client.name(), err)
	}
	return values, nil
}

// Store and if the secret already exists, update it.
func (a *ASM) Store(ctx context.Context, ref configuration.Ref, value []byte) (*url.URL, error) {
	url, err := a.client.store(ctx, ref, value)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", a.client.name(), err)
	}
	return url, nil
}

func (a *ASM) Delete(ctx context.Context, ref configuration.Ref) error {
	err := a.client.delete(ctx, ref)
	if err != nil {
		return fmt.Errorf("%s: %w", a.client.name(), err)
	}
	return nil
}
