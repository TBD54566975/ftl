package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"connectrpc.com/connect"
	"github.com/alecthomas/kong"
	_ "github.com/jackc/pgx/v5/stdlib"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/common/encoding"
	"github.com/TBD54566975/ftl/common/schema"
	"github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/configuration/manager"
	"github.com/TBD54566975/ftl/internal/configuration/providers"
	"github.com/TBD54566975/ftl/internal/dsn"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
	"github.com/TBD54566975/ftl/internal/schema/schemaeventsource"
)

type Config struct {
	Bind *url.URL `help:"Socket to bind to." default:"http://127.0.0.1:8896" env:"FTL_BIND"`
}

func (c *Config) SetDefaults() {
	if err := kong.ApplyDefaults(c, kong.Vars{"dsn": dsn.PostgresDSN("ftl")}); err != nil {
		panic(err)
	}
}

type AdminService struct {
	schr SchemaRetriever
	cm   *manager.Manager[configuration.Configuration]
	sm   *manager.Manager[configuration.Secrets]
}

var _ ftlv1connect.AdminServiceHandler = (*AdminService)(nil)

type SchemaRetriever interface {
	// BindAllocator is required if the schema is retrieved from disk using language plugins
	GetActiveSchema(ctx context.Context) (*schema.Schema, error)
}

func NewSchemaRetreiver(source schemaeventsource.EventSource) SchemaRetriever {
	return &streamSchemaRetriever{
		source: source,
	}
}

type streamSchemaRetriever struct {
	source schemaeventsource.EventSource
}

func (c streamSchemaRetriever) GetActiveSchema(ctx context.Context) (*schema.Schema, error) {
	view := c.source.View()
	return &schema.Schema{Modules: view.Modules}, nil
}

// NewAdminService creates a new AdminService.
// bindAllocator is optional and should be set if a local client is to be used that accesses schema from disk using language plugins.
func NewAdminService(cm *manager.Manager[configuration.Configuration], sm *manager.Manager[configuration.Secrets], schr SchemaRetriever) *AdminService {
	return &AdminService{
		schr: schr,
		cm:   cm,
		sm:   sm,
	}
}

func Start(
	ctx context.Context,
	config Config,
	cm *manager.Manager[configuration.Configuration],
	sm *manager.Manager[configuration.Secrets],
	schr SchemaRetriever,
) error {
	config.SetDefaults()

	logger := log.FromContext(ctx).Scope("admin")
	svc := NewAdminService(cm, sm, schr)

	logger.Debugf("Admin service listening on: %s", config.Bind)
	err := rpc.Serve(ctx, config.Bind,
		rpc.GRPC(ftlv1connect.NewAdminServiceHandler, svc),
	)
	if err != nil {
		return fmt.Errorf("admin service stopped serving: %w", err)
	}
	return nil
}

func (s *AdminService) Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error) {
	return connect.NewResponse(&ftlv1.PingResponse{}), nil
}

// ConfigList returns the list of configuration values, optionally filtered by module.
func (s *AdminService) ConfigList(ctx context.Context, req *connect.Request[ftlv1.ConfigListRequest]) (*connect.Response[ftlv1.ConfigListResponse], error) {
	listing, err := s.cm.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list configs: %w", err)
	}

	configs := []*ftlv1.ConfigListResponse_Config{}
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

		configs = append(configs, &ftlv1.ConfigListResponse_Config{
			RefPath: ref,
			Value:   cv,
		})
	}
	return connect.NewResponse(&ftlv1.ConfigListResponse{Configs: configs}), nil
}

// ConfigGet returns the configuration value for a given ref string.
func (s *AdminService) ConfigGet(ctx context.Context, req *connect.Request[ftlv1.ConfigGetRequest]) (*connect.Response[ftlv1.ConfigGetResponse], error) {
	var value any
	err := s.cm.Get(ctx, refFromConfigRef(req.Msg.GetRef()), &value)
	if err != nil {
		return nil, fmt.Errorf("failed to get from config manager: %w", err)
	}
	vb, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
	}
	return connect.NewResponse(&ftlv1.ConfigGetResponse{Value: vb}), nil
}

func configProviderKey(p *ftlv1.ConfigProvider) configuration.ProviderKey {
	if p == nil {
		return ""
	}
	switch *p {
	case ftlv1.ConfigProvider_CONFIG_PROVIDER_INLINE:
		return providers.InlineProviderKey
	case ftlv1.ConfigProvider_CONFIG_PROVIDER_ENVAR:
		return providers.EnvarProviderKey
	}
	return ""
}

// ConfigSet sets the configuration at the given ref to the provided value.
func (s *AdminService) ConfigSet(ctx context.Context, req *connect.Request[ftlv1.ConfigSetRequest]) (*connect.Response[ftlv1.ConfigSetResponse], error) {
	err := s.validateAgainstSchema(ctx, false, refFromConfigRef(req.Msg.GetRef()), req.Msg.Value)
	if err != nil {
		return nil, err
	}

	err = s.cm.SetJSON(ctx, refFromConfigRef(req.Msg.GetRef()), req.Msg.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to set config: %w", err)
	}
	return connect.NewResponse(&ftlv1.ConfigSetResponse{}), nil
}

// ConfigUnset unsets the config value at the given ref.
func (s *AdminService) ConfigUnset(ctx context.Context, req *connect.Request[ftlv1.ConfigUnsetRequest]) (*connect.Response[ftlv1.ConfigUnsetResponse], error) {
	err := s.cm.Unset(ctx, refFromConfigRef(req.Msg.GetRef()))
	if err != nil {
		return nil, fmt.Errorf("failed to unset config: %w", err)
	}
	return connect.NewResponse(&ftlv1.ConfigUnsetResponse{}), nil
}

// SecretsList returns the list of secrets, optionally filtered by module.
func (s *AdminService) SecretsList(ctx context.Context, req *connect.Request[ftlv1.SecretsListRequest]) (*connect.Response[ftlv1.SecretsListResponse], error) {
	listing, err := s.sm.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}
	secrets := []*ftlv1.SecretsListResponse_Secret{}
	for _, secret := range listing {
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
		secrets = append(secrets, &ftlv1.SecretsListResponse_Secret{
			RefPath: ref,
			Value:   sv,
		})
	}
	return connect.NewResponse(&ftlv1.SecretsListResponse{Secrets: secrets}), nil
}

// SecretGet returns the secret value for a given ref string.
func (s *AdminService) SecretGet(ctx context.Context, req *connect.Request[ftlv1.SecretGetRequest]) (*connect.Response[ftlv1.SecretGetResponse], error) {
	var value any
	err := s.sm.Get(ctx, refFromConfigRef(req.Msg.GetRef()), &value)
	if err != nil {
		return nil, fmt.Errorf("failed to get from secret manager: %w", err)
	}
	vb, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
	}
	return connect.NewResponse(&ftlv1.SecretGetResponse{Value: vb}), nil
}

// SecretSet sets the secret at the given ref to the provided value.
func (s *AdminService) SecretSet(ctx context.Context, req *connect.Request[ftlv1.SecretSetRequest]) (*connect.Response[ftlv1.SecretSetResponse], error) {
	err := s.validateAgainstSchema(ctx, true, refFromConfigRef(req.Msg.GetRef()), req.Msg.Value)
	if err != nil {
		return nil, err
	}

	err = s.sm.SetJSON(ctx, refFromConfigRef(req.Msg.GetRef()), req.Msg.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to set secret: %w", err)
	}
	return connect.NewResponse(&ftlv1.SecretSetResponse{}), nil
}

// SecretUnset unsets the secret value at the given ref.
func (s *AdminService) SecretUnset(ctx context.Context, req *connect.Request[ftlv1.SecretUnsetRequest]) (*connect.Response[ftlv1.SecretUnsetResponse], error) {
	err := s.sm.Unset(ctx, refFromConfigRef(req.Msg.GetRef()))
	if err != nil {
		return nil, fmt.Errorf("failed to unset secret: %w", err)
	}
	return connect.NewResponse(&ftlv1.SecretUnsetResponse{}), nil
}

// MapConfigsForModule combines all configuration values visible to the module.
func (s *AdminService) MapConfigsForModule(ctx context.Context, req *connect.Request[ftlv1.MapConfigsForModuleRequest]) (*connect.Response[ftlv1.MapConfigsForModuleResponse], error) {
	values, err := s.cm.MapForModule(ctx, req.Msg.Module)
	if err != nil {
		return nil, fmt.Errorf("failed to map configs for module: %w", err)
	}
	return connect.NewResponse(&ftlv1.MapConfigsForModuleResponse{Values: values}), nil
}

// MapSecretsForModule combines all secrets visible to the module.
func (s *AdminService) MapSecretsForModule(ctx context.Context, req *connect.Request[ftlv1.MapSecretsForModuleRequest]) (*connect.Response[ftlv1.MapSecretsForModuleResponse], error) {
	values, err := s.sm.MapForModule(ctx, req.Msg.Module)
	if err != nil {
		return nil, fmt.Errorf("failed to map secrets for module: %w", err)
	}
	return connect.NewResponse(&ftlv1.MapSecretsForModuleResponse{Values: values}), nil
}

func refFromConfigRef(cr *ftlv1.ConfigRef) configuration.Ref {
	return configuration.NewRef(cr.GetModule(), cr.GetName())
}

func (s *AdminService) validateAgainstSchema(ctx context.Context, isSecret bool, ref configuration.Ref, value json.RawMessage) error {
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
		logger.Debugf("skipping validation; declaration %q not found", ref.Name)
		return nil
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
