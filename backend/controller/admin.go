package controller

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
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list config: %w", err))
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
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get value: %w", err))
			} else {
				cv, _ = json.Marshal(value)
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
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get value: %w", err))
	}

	vb, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to encode value: %w", err))
	}

	return connect.NewResponse(&ftlv1.GetConfigResponse{Value: vb}), nil
}

// ConfigSet sets the configuration at the given ref to the provided value.
func (s *AdminService) ConfigSet(ctx context.Context, req *connect.Request[ftlv1.SetConfigRequest]) (*connect.Response[ftlv1.SetConfigResponse], error) {
	var err error

	// TODO(saf): use req.Msg.Provider to update / create a manager with the correct provider
	cm := s.cm

	if err := cm.Mutable(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	err = cm.Set(ctx, cf.NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name), req.Msg.Value)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to set value: %w", err))
	}

	// TODO(saf) save the updated config

	return connect.NewResponse(&ftlv1.SetConfigResponse{}), nil
}

// ConfigUnset unsets the config value at the given ref.
func (s *AdminService) ConfigUnset(ctx context.Context, req *connect.Request[ftlv1.UnsetConfigRequest]) (*connect.Response[ftlv1.UnsetConfigResponse], error) {
	err := s.cm.Unset(ctx, cf.NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to unset value: %w", err))
	}
	return connect.NewResponse(&ftlv1.UnsetConfigResponse{}), nil
}

// SecretsList returns the list of secrets, optionally filtered by module.
func (s *AdminService) SecretsList(ctx context.Context, req *connect.Request[ftlv1.ListSecretsRequest]) (*connect.Response[ftlv1.ListSecretsResponse], error) {
	return connect.NewResponse(&ftlv1.ListSecretsResponse{}), nil
}

// SecretGet returns the secret value for a given ref string.
func (s *AdminService) SecretGet(ctx context.Context, req *connect.Request[ftlv1.GetSecretRequest]) (*connect.Response[ftlv1.GetSecretResponse], error) {
	return connect.NewResponse(&ftlv1.GetSecretResponse{}), nil
}

// SecretSet sets the secret at the given ref to the provided value.
func (s *AdminService) SecretSet(ctx context.Context, req *connect.Request[ftlv1.SetSecretRequest]) (*connect.Response[ftlv1.SetSecretResponse], error) {
	return connect.NewResponse(&ftlv1.SetSecretResponse{}), nil
}

// SecretUnset unsets the secret value at the given ref.
func (s *AdminService) SecretUnset(ctx context.Context, req *connect.Request[ftlv1.UnsetSecretRequest]) (*connect.Response[ftlv1.UnsetSecretResponse], error) {
	return connect.NewResponse(&ftlv1.UnsetSecretResponse{}), nil
}
