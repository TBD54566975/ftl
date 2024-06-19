package admin

import (
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	cf "github.com/TBD54566975/ftl/common/configuration"
)

type AdminService struct {
	cm *cf.Manager[cf.Configuration]
	sm *cf.Manager[cf.Secrets]
}

var _ ftlv1connect.AdminServiceHandler = (*AdminService)(nil)

func NewAdminService(cm *cf.Manager[cf.Configuration], sm *cf.Manager[cf.Secrets]) *AdminService {
	return &AdminService{
		cm: cm,
		sm: sm,
	}
}

func (s *AdminService) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

// ConfigList returns the list of configuration values, optionally filtered by module.
func (s *AdminService) ConfigList(ctx context.Context, req *connect.Request[ftlv1.ListConfigRequest]) (*connect.Response[ftlv1.ListConfigResponse], error) {
	listing, err := s.cm.List(ctx)
	if err != nil {
		return nil, err
	}

	configs := []*ftlv1.ListConfigResponse_Config{}
	for _, config := range listing {
		module, ok := config.Module.Get()
		if *req.Msg.Module != "" && module != *req.Msg.Module {
			continue
		}

		ref := config.Name
		if ok {
			ref = fmt.Sprintf("%s.%s", module, config.Name)
		}

		var cv []byte
		if *req.Msg.IncludeValues {
			var value any
			err := s.cm.Get(ctx, config.Ref, &value)
			if err != nil {
				return nil, err
			}
			cv, err = json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal value for %s: %w", ref, err)
			}
		}

		configs = append(configs, &ftlv1.ListConfigResponse_Config{
			RefPath: ref,
			Value:   cv,
		})
	}
	return connect.NewResponse(&ftlv1.ListConfigResponse{Configs: configs}), nil
}

// ConfigGet returns the configuration value for a given ref string.
func (s *AdminService) ConfigGet(ctx context.Context, req *connect.Request[ftlv1.GetConfigRequest]) (*connect.Response[ftlv1.GetConfigResponse], error) {
	var value any
	err := s.cm.Get(ctx, cf.NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name), &value)
	if err != nil {
		return nil, err
	}
	vb, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.GetConfigResponse{Value: vb}), nil
}

func configProviderKey(p *ftlv1.ConfigProvider) string {
	if p == nil {
		return ""
	}
	switch *p {
	case ftlv1.ConfigProvider_CONFIG_INLINE:
		return "inline"
	case ftlv1.ConfigProvider_CONFIG_ENVAR:
		return "envar"
	case ftlv1.ConfigProvider_CONFIG_DB:
		return "db"
	}
	return ""
}

// ConfigSet sets the configuration at the given ref to the provided value.
func (s *AdminService) ConfigSet(ctx context.Context, req *connect.Request[ftlv1.SetConfigRequest]) (*connect.Response[ftlv1.SetConfigResponse], error) {
	pkey := configProviderKey(req.Msg.Provider)
	err := s.cm.SetJSON(ctx, pkey, cf.NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name), req.Msg.Value)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.SetConfigResponse{}), nil
}

// ConfigUnset unsets the config value at the given ref.
func (s *AdminService) ConfigUnset(ctx context.Context, req *connect.Request[ftlv1.UnsetConfigRequest]) (*connect.Response[ftlv1.UnsetConfigResponse], error) {
	pkey := configProviderKey(req.Msg.Provider)
	err := s.cm.Unset(ctx, pkey, cf.NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.UnsetConfigResponse{}), nil
}

// SecretsList returns the list of secrets, optionally filtered by module.
func (s *AdminService) SecretsList(ctx context.Context, req *connect.Request[ftlv1.ListSecretsRequest]) (*connect.Response[ftlv1.ListSecretsResponse], error) {
	listing, err := s.sm.List(ctx)
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
			var value any
			err := s.sm.Get(ctx, secret.Ref, &value)
			if err != nil {
				return nil, err
			}
			sv, err = json.Marshal(value)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal value for %s: %w", ref, err)
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
func (s *AdminService) SecretGet(ctx context.Context, req *connect.Request[ftlv1.GetSecretRequest]) (*connect.Response[ftlv1.GetSecretResponse], error) {
	var value any
	err := s.sm.Get(ctx, cf.NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name), &value)
	if err != nil {
		return nil, err
	}
	vb, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.GetSecretResponse{Value: vb}), nil
}

func secretProviderKey(p *ftlv1.SecretProvider) string {
	if p == nil {
		return ""
	}
	switch *p {
	case ftlv1.SecretProvider_SECRET_INLINE:
		return "inline"
	case ftlv1.SecretProvider_SECRET_ENVAR:
		return "envar"
	case ftlv1.SecretProvider_SECRET_KEYCHAIN:
		return "keychain"
	case ftlv1.SecretProvider_SECRET_OP:
		return "op"
	case ftlv1.SecretProvider_SECRET_ASM:
		return "asm"
	}
	return ""
}

// SecretSet sets the secret at the given ref to the provided value.
func (s *AdminService) SecretSet(ctx context.Context, req *connect.Request[ftlv1.SetSecretRequest]) (*connect.Response[ftlv1.SetSecretResponse], error) {
	pkey := secretProviderKey(req.Msg.Provider)
	err := s.sm.SetJSON(ctx, pkey, cf.NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name), req.Msg.Value)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.SetSecretResponse{}), nil
}

// SecretUnset unsets the secret value at the given ref.
func (s *AdminService) SecretUnset(ctx context.Context, req *connect.Request[ftlv1.UnsetSecretRequest]) (*connect.Response[ftlv1.UnsetSecretResponse], error) {
	pkey := secretProviderKey(req.Msg.Provider)
	err := s.sm.Unset(ctx, pkey, cf.NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name))
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&ftlv1.UnsetSecretResponse{}), nil
}
