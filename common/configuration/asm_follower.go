package configuration

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/alecthomas/types/optional"
)

// asmFollower uses AdminService to get/set secrets from the leader
type asmFollower struct {
	client ftlv1connect.AdminServiceClient
}

var _ asmClient = &asmFollower{}

func (f *asmFollower) list(ctx context.Context) ([]Entry, error) {
	module := ""
	includeValues := false
	resp, err := f.client.SecretsList(ctx, connect.NewRequest(&ftlv1.ListSecretsRequest{
		Module:        &module,
		IncludeValues: &includeValues,
	}))
	if err != nil {
		return nil, err
	}
	entries := []Entry{}
	for _, s := range resp.Msg.Secrets {
		components := strings.Split(s.RefPath, ".")
		var ref Ref
		switch len(components) {
		case 1:
			ref = Ref{
				Name: components[0],
			}
		case 2:
			ref = Ref{
				Module: optional.Some(components[0]),
				Name:   components[1],
			}
		default:
			return nil, fmt.Errorf("invalid ref path: %s", s.RefPath)
		}
		entries = append(entries, Entry{
			Ref:      ref,
			Accessor: asmURLForRef(ref),
		})
	}

	return entries, nil
}

func (f *asmFollower) load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	resp, err := f.client.SecretGet(ctx, connect.NewRequest(&ftlv1.GetSecretRequest{
		Ref: &ftlv1.ConfigRef{
			Module: ref.Module.Ptr(),
			Name:   ref.Name,
		},
	}))
	if err != nil {
		return nil, err
	}
	return resp.Msg.Value, nil
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
	return err
}
