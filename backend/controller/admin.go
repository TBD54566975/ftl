package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/internal/log"
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

// List configuration.
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

		cv := ""
		if *req.Msg.IncludeValues {
			var value any
			err := s.cm.Get(ctx, config.Ref, &value)
			if err != nil {
				return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get value: %w", err))
			} else {
				data, _ := json.Marshal(value)
				cv = string(data)
			}
		}

		configs = append(configs, &ftlv1.ListConfigResponse_Config{
			RefPath: ref,
			Value:   &cv,
		})
	}
	return connect.NewResponse(&ftlv1.ListConfigResponse{Configs: configs}), nil
}

// Get a config value.
func (s *AdminService) ConfigGet(ctx context.Context, req *connect.Request[ftlv1.GetConfigRequest]) (*connect.Response[ftlv1.GetConfigResponse], error) {
	var value any
	err := s.cm.Get(ctx, cf.NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name), &value)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get value: %w", err))
	}

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	err = enc.Encode(value)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to encode value: %w", err))
	}

	return connect.NewResponse(&ftlv1.GetConfigResponse{Value: buf.String()}), nil
}

// Set a config value.
func (s *AdminService) ConfigSet(ctx context.Context, req *connect.Request[ftlv1.SetConfigRequest]) (*connect.Response[ftlv1.SetConfigResponse], error) {
	// cm := cf.ConfigFromContext(ctx) // TODO(saf): use cf.New to create a cm with the appropriate provider/writer

	var err error

	// var providerKey string
	// switch *req.Msg.Provider {
	// case ftlv1.ConfigProvider_CONFIG_ENVAR:
	// 	providerKey = "envar"
	// case ftlv1.ConfigProvider_CONFIG_INLINE:
	// 	providerKey = "inline"
	// }

	logger := log.FromContext(ctx)
	logger.Warnf("cm pre-mutable %+v", s.cm)
	if err := s.cm.Mutable(); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	logger.Warnf("cm post-mutable %+v", s.cm)

	err = s.cm.Set(ctx, cf.NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name), req.Msg.Value)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to set value: %w", err))
	}

	// TODO save the updated config

	return connect.NewResponse(&ftlv1.SetConfigResponse{}), nil
}

// Unset a config value.
func (s *AdminService) ConfigUnset(ctx context.Context, req *connect.Request[ftlv1.UnsetConfigRequest]) (*connect.Response[ftlv1.UnsetConfigResponse], error) {
	err := s.cm.Unset(ctx, cf.NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name))
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to unset value: %w", err))
	}
	return connect.NewResponse(&ftlv1.UnsetConfigResponse{}), nil
}

// List secrets.
func (s *AdminService) SecretsList(ctx context.Context, req *connect.Request[ftlv1.ListSecretsRequest]) (*connect.Response[ftlv1.ListSecretsResponse], error) {
	return connect.NewResponse(&ftlv1.ListSecretsResponse{}), nil
}

// Get a secret.
func (s *AdminService) SecretGet(ctx context.Context, req *connect.Request[ftlv1.GetSecretRequest]) (*connect.Response[ftlv1.GetSecretResponse], error) {
	return connect.NewResponse(&ftlv1.GetSecretResponse{}), nil
}

// Set a secret.
func (s *AdminService) SecretSet(ctx context.Context, req *connect.Request[ftlv1.SetSecretRequest]) (*connect.Response[ftlv1.SetSecretResponse], error) {
	return connect.NewResponse(&ftlv1.SetSecretResponse{}), nil
}

// Unset a secret.
func (s *AdminService) SecretUnset(ctx context.Context, req *connect.Request[ftlv1.UnsetSecretRequest]) (*connect.Response[ftlv1.UnsetSecretResponse], error) {
	return connect.NewResponse(&ftlv1.UnsetSecretResponse{}), nil
}
