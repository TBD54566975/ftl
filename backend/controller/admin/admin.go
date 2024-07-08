package admin

import (
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/schema"
	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/go-runtime/encoding"
	"github.com/TBD54566975/ftl/internal/log"
)

type AdminService struct {
	schr SchemaRetriever
	cm   *cf.Manager[cf.Configuration]
	sm   *cf.Manager[cf.Secrets]
}

var _ ftlv1connect.AdminServiceHandler = (*AdminService)(nil)

type SchemaRetriever interface {
	GetActiveSchema(ctx context.Context) (*schema.Schema, error)
}

func NewAdminService(cm *cf.Manager[cf.Configuration], sm *cf.Manager[cf.Secrets], schr SchemaRetriever) *AdminService {
	return &AdminService{
		schr: schr,
		cm:   cm,
		sm:   sm,
	}
}

func (s *AdminService) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

// ConfigList returns the list of configuration values, optionally filtered by module.
func (s *AdminService) ConfigList(ctx context.Context, req *connect.Request[ftlv1.ListConfigRequest]) (*connect.Response[ftlv1.ListConfigResponse], error) {
	listing, err := s.cm.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list configs: %w", err)
	}

	configs := []*ftlv1.ListConfigResponse_Config{}
	for _, config := range listing {
		module, ok := config.Module.Get()
		if req.Msg.Module != nil && *req.Msg.Module != "" && module != *req.Msg.Module {
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
				return nil, fmt.Errorf("failed to get value for %v: %w", ref, err)
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
	err := s.cm.Get(ctx, refFromConfigRef(req.Msg.GetRef()), &value)
	if err != nil {
		return nil, fmt.Errorf("failed to get from config manager: %w", err)
	}
	vb, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
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
	err := s.validateAgainstSchema(ctx, false, refFromConfigRef(req.Msg.GetRef()), req.Msg.Value)
	if err != nil {
		return nil, err
	}

	pkey := configProviderKey(req.Msg.Provider)
	err = s.cm.SetJSON(ctx, pkey, refFromConfigRef(req.Msg.GetRef()), req.Msg.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to set config: %w", err)
	}
	return connect.NewResponse(&ftlv1.SetConfigResponse{}), nil
}

// ConfigUnset unsets the config value at the given ref.
func (s *AdminService) ConfigUnset(ctx context.Context, req *connect.Request[ftlv1.UnsetConfigRequest]) (*connect.Response[ftlv1.UnsetConfigResponse], error) {
	pkey := configProviderKey(req.Msg.Provider)
	err := s.cm.Unset(ctx, pkey, refFromConfigRef(req.Msg.GetRef()))
	if err != nil {
		return nil, fmt.Errorf("failed to unset config: %w", err)
	}
	return connect.NewResponse(&ftlv1.UnsetConfigResponse{}), nil
}

// SecretsList returns the list of secrets, optionally filtered by module.
func (s *AdminService) SecretsList(ctx context.Context, req *connect.Request[ftlv1.ListSecretsRequest]) (*connect.Response[ftlv1.ListSecretsResponse], error) {
	listing, err := s.sm.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}
	secrets := []*ftlv1.ListSecretsResponse_Secret{}
	for _, secret := range listing {
		if req.Msg.Provider != nil && cf.ProviderKeyForAccessor(secret.Accessor) != secretProviderKey(req.Msg.Provider) {
			// Skip secrets that don't match the provider in the request
			continue
		}
		module, ok := secret.Module.Get()
		if req.Msg.Module != nil && *req.Msg.Module != "" && module != *req.Msg.Module {
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
				return nil, fmt.Errorf("failed to get value for %v: %w", ref, err)
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
	err := s.sm.Get(ctx, refFromConfigRef(req.Msg.GetRef()), &value)
	if err != nil {
		return nil, fmt.Errorf("failed to get from secret manager: %w", err)
	}
	vb, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
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
	err := s.validateAgainstSchema(ctx, true, refFromConfigRef(req.Msg.GetRef()), req.Msg.Value)
	if err != nil {
		return nil, err
	}

	pkey := secretProviderKey(req.Msg.Provider)
	err = s.sm.SetJSON(ctx, pkey, refFromConfigRef(req.Msg.GetRef()), req.Msg.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to set secret: %w", err)
	}
	return connect.NewResponse(&ftlv1.SetSecretResponse{}), nil
}

// SecretUnset unsets the secret value at the given ref.
func (s *AdminService) SecretUnset(ctx context.Context, req *connect.Request[ftlv1.UnsetSecretRequest]) (*connect.Response[ftlv1.UnsetSecretResponse], error) {
	pkey := secretProviderKey(req.Msg.Provider)
	err := s.sm.Unset(ctx, pkey, refFromConfigRef(req.Msg.GetRef()))
	if err != nil {
		return nil, fmt.Errorf("failed to unset secret: %w", err)
	}
	return connect.NewResponse(&ftlv1.UnsetSecretResponse{}), nil
}

func refFromConfigRef(cr *ftlv1.ConfigRef) cf.Ref {
	return cf.NewRef(cr.GetModule(), cr.GetName())
}

func (s *AdminService) validateAgainstSchema(ctx context.Context, isSecret bool, ref cf.Ref, value json.RawMessage) error {
	logger := log.FromContext(ctx)

	// Globals aren't in the module schemas, so we have nothing to validate against.
	if !ref.Module.Ok() {
		return nil
	}

	// If we can't retrieve an active schema, skip validation.
	sch, err := s.schr.GetActiveSchema(ctx)
	if err != nil {
		logger.Debugf("skipping validation; could not get the active schema: %v", err)
		return nil
	}

	r := schema.RefKey{Module: ref.Module.Default(""), Name: ref.Name}.ToRef()
	decl, ok := sch.Resolve(r).Get()
	if !ok {
		return fmt.Errorf("declaration %q not found", ref.Name)
	}

	var fieldType schema.Type
	if isSecret {
		decl, ok := decl.(*schema.Secret)
		if !ok {
			return fmt.Errorf("%q is not a secret declaration", ref.Name)
		}
		fieldType = decl.Type
	} else {
		decl, ok := decl.(*schema.Config)
		if !ok {
			return fmt.Errorf("%q is not a config declaration", ref.Name)
		}
		fieldType = decl.Type
	}

	var v any
	err = encoding.Unmarshal(value, &v)
	if err != nil {
		return fmt.Errorf("could not unmarshal JSON value: %w", err)
	}

	err = schema.ValidateJSONValue(fieldType, []string{ref.Name}, v, sch)
	if err != nil {
		return fmt.Errorf("JSON validation failed: %w", err)
	}

	return nil
}
