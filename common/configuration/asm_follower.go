package configuration

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"connectrpc.com/connect"
	"github.com/puzpuzpuz/xsync/v3"

	"github.com/TBD54566975/ftl/backend/controller/leader"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/log"
)

const asmFollowerSyncInterval = time.Second * 10

// asmFollower uses AdminService to get/set secrets from the leader
type asmFollower struct {
	errorFilter *leader.ErrorFilter
	leaderName  string
	// client requests/responses use unobfuscated values
	client ftlv1connect.AdminServiceClient
}

func newASMFollower(rpcClient ftlv1connect.AdminServiceClient, leaderName string, leaseTTL time.Duration) *asmFollower {
	f := &asmFollower{
		errorFilter: leader.NewErrorFilter(leaseTTL),
		leaderName:  leaderName,
		client:      rpcClient,
	}
	return f
}

func (f *asmFollower) name() string {
	return fmt.Sprintf("asm/follower/%s", f.leaderName)
}

func (f *asmFollower) syncInterval() time.Duration {
	return asmFollowerSyncInterval
}

func (f *asmFollower) sync(ctx context.Context, values *xsync.MapOf[Ref, SyncedValue]) error {
	// values must store obfuscated values, but f.client gives unobfuscated values
	logger := log.FromContext(ctx)
	obfuscator := Secrets{}.obfuscator()
	module := ""
	includeValues := true
	resp, err := f.client.SecretsList(ctx, connect.NewRequest(&ftlv1.ListSecretsRequest{
		Module:        &module,
		IncludeValues: &includeValues,
	}))
	if err != nil {
		if connectErr := new(connect.Error); errors.As(err, &connectErr) {
			if connectErr.Code() == connect.CodeInternal || connectErr.Code() == connect.CodeUnavailable {
				if !f.errorFilter.ReportLeaseError() {
					logger.Warnf("error getting secrets list from leader, possible leader failover %s", err.Error())
					return nil
				}
			}
		}
		return fmt.Errorf("error getting secrets list from leader: %w", err)
	}
	f.errorFilter.ReportOperationSuccess()
	visited := map[Ref]bool{}
	for _, s := range resp.Msg.Secrets {
		ref, err := ParseRef(s.RefPath)
		if err != nil {
			return fmt.Errorf("invalid ref %q: %w", s.RefPath, err)
		}
		obfuscatedValue, err := obfuscator.Obfuscate(s.Value)
		if err != nil {
			return fmt.Errorf("asm follower could not obfuscate value for ref %q: %w", s.RefPath, err)
		}
		visited[ref] = true
		values.Store(ref, SyncedValue{
			Value: obfuscatedValue,
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

func (f *asmFollower) store(ctx context.Context, ref Ref, obfuscatedValue []byte) (*url.URL, error) {
	obfuscator := Secrets{}.obfuscator()
	unobfuscatedValue, err := obfuscator.Reveal(obfuscatedValue)
	if err != nil {
		return nil, fmt.Errorf("asm follower could not unobfuscate: %w", err)
	}
	provider := ftlv1.SecretProvider_SECRET_ASM
	_, err = f.client.SecretSet(ctx, connect.NewRequest(&ftlv1.SetSecretRequest{
		Provider: &provider,
		Ref: &ftlv1.ConfigRef{
			Module: ref.Module.Ptr(),
			Name:   ref.Name,
		},
		Value: unobfuscatedValue,
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
