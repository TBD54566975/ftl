package configuration

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/puzpuzpuz/xsync/v3"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
)

const asmFollowerSyncInterval = time.Minute * 1

// asmFollower uses AdminService to get/set secrets from the leader
type asmFollower struct {
	client ftlv1connect.AdminServiceClient
}

var _ asmClient = &asmFollower{}

func newASMFollower(ctx context.Context, rpcClient ftlv1connect.AdminServiceClient) *asmFollower {
	f := &asmFollower{
		client: rpcClient,
	}
	return f
}

func (f *asmFollower) syncInterval() time.Duration {
	return asmFollowerSyncInterval
}

func (f *asmFollower) sync(ctx context.Context, values *xsync.MapOf[Ref, SyncedValue]) error {
	module := ""
	includeValues := true
	resp, err := f.client.SecretsList(ctx, connect.NewRequest(&ftlv1.ListSecretsRequest{
		Module:        &module,
		IncludeValues: &includeValues,
	}))
	if err != nil {
		return fmt.Errorf("error getting secrets list from leader: %w", err)
	}
	visited := map[Ref]bool{}
	for _, s := range resp.Msg.Secrets {
		ref, err := ParseRef(s.RefPath)
		if err != nil {
			return fmt.Errorf("invalid ref %q: %w", s.RefPath, err)
		}
		visited[ref] = true
		values.Store(ref, SyncedValue{
			Value: s.Value,
		})
	}
	// delete old values
	values.Range(func(ref Ref, _ SyncedValue) bool {
		if !visited[ref] {
			values.Delete(ref)
		}
		return true
	})
	return nil
}

func (f *asmFollower) store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	provider := ftlv1.SecretProvider_SECRET_ASM
	_, err := f.client.SecretSet(ctx, connect.NewRequest(&ftlv1.SetSecretRequest{
		Provider: &provider,
		Ref: &ftlv1.ConfigRef{
			Module: ref.Module.Ptr(),
			Name:   ref.Name,
		},
		Value: value,
	}))
	if err != nil {
		return nil, err
	}
	return asmURLForRef(ref), nil
}

func (f *asmFollower) delete(ctx context.Context, ref Ref) error {
	provider := ftlv1.SecretProvider_SECRET_ASM
	_, err := f.client.SecretUnset(ctx, connect.NewRequest(&ftlv1.UnsetSecretRequest{
		Provider: &provider,
		Ref: &ftlv1.ConfigRef{
			Module: ref.Module.Ptr(),
			Name:   ref.Name,
		},
	}))
	if err != nil {
		return err
	}
	return nil
}
