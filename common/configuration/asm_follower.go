package configuration

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/puzpuzpuz/xsync/v3"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
)

const asmFollowerSyncInterval = time.Minute * 1

// asmFollower uses AdminService to get/set secrets from the leader
type asmFollower struct {
	// client requests/responses use unobfuscated values
	client ftlv1connect.AdminServiceClient

	// cache stores obfuscated values
	cache *secretsCache
}

var _ asmClient = &asmFollower{}

func newASMFollower(ctx context.Context, rpcClient ftlv1connect.AdminServiceClient, clock clock.Clock) *asmFollower {
	f := &asmFollower{
		client: rpcClient,
		cache:  newSecretsCache("asm-follower"),
	}
	go f.cache.sync(ctx, asmFollowerSyncInterval, func(ctx context.Context, secrets *xsync.MapOf[Ref, cachedSecret]) error {
		return f.sync(ctx, secrets)
	}, clock)
	return f
}

func (f *asmFollower) sync(ctx context.Context, secrets *xsync.MapOf[Ref, cachedSecret]) error {
	obfuscator := Secrets{}.obfuscator()
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
		obfuscatedValue, err := obfuscator.Obfuscate(s.Value)
		if err != nil {
			return fmt.Errorf("asm follower could not obfuscate value for ref %q: %w", s.RefPath, err)
		}
		visited[ref] = true
		secrets.Store(ref, cachedSecret{
			value: obfuscatedValue,
		})
	}
	// delete old values
	secrets.Range(func(ref Ref, _ cachedSecret) bool {
		if !visited[ref] {
			secrets.Delete(ref)
		}
		return true
	})
	return nil
}

// list all secrets in the account.
func (f *asmFollower) list(ctx context.Context) ([]Entry, error) {
	entries := []Entry{}
	err := f.cache.iterate(func(ref Ref, _ []byte) {
		entries = append(entries, Entry{
			Ref:      ref,
			Accessor: asmURLForRef(ref),
		})
	})
	if err != nil {
		return nil, err
	}
	return entries, nil
}

func (f *asmFollower) load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	return f.cache.getSecret(ref)
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
	f.cache.updatedSecret(ref, obfuscatedValue)
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
	f.cache.deletedSecret(ref)
	return nil
}
