package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/backend/controller/dal"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/buildengine"
	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/common/projectconfig"
	"github.com/TBD54566975/ftl/internal/slices"
)

type AdminService struct {
	dal optional.Option[*dal.DAL]
	cm  *cf.Manager[cf.Configuration]
	sm  *cf.Manager[cf.Secrets]
}

var _ ftlv1connect.AdminServiceHandler = (*AdminService)(nil)

func NewAdminService(cm *cf.Manager[cf.Configuration], sm *cf.Manager[cf.Secrets], dal optional.Option[*dal.DAL]) *AdminService {
	return &AdminService{
		dal: dal,
		cm:  cm,
		sm:  sm,
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
	err := validateAgainstSchema(ctx, s.dal, false, cf.NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name), req.Msg.Value)
	if err != nil {
		return nil, err
	}

	pkey := configProviderKey(req.Msg.Provider)
	err = s.cm.SetJSON(ctx, pkey, cf.NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name), req.Msg.Value)
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
	err := validateAgainstSchema(ctx, s.dal, true, cf.NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name), req.Msg.Value)
	if err != nil {
		return nil, err
	}

	pkey := secretProviderKey(req.Msg.Provider)
	err = s.sm.SetJSON(ctx, pkey, cf.NewRef(*req.Msg.Ref.Module, req.Msg.Ref.Name), req.Msg.Value)
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

func schemaFromDal(ctx context.Context, optDal optional.Option[*dal.DAL]) (*schema.Schema, error) {
	d, ok := optDal.Get()
	if !ok {
		return nil, fmt.Errorf("no DAL available")
	}
	deployments, err := d.GetActiveDeployments(ctx)
	if err != nil {
		return nil, err
	}
	return schema.ValidateSchema(&schema.Schema{
		Modules: slices.Map(deployments, func(depl dal.Deployment) *schema.Module {
			return depl.Schema
		}),
	})
}

func schemaFromDisk(ctx context.Context) (*schema.Schema, error) {
	path, ok := projectconfig.DefaultConfigPath().Get()
	if !ok {
		return nil, fmt.Errorf("no project config path available")
	}
	projConfig, err := projectconfig.Load(ctx, path)
	if err != nil {
		return nil, err
	}
	modules, err := buildengine.DiscoverModules(ctx, projConfig.AbsModuleDirs())
	if err != nil {
		return nil, err
	}

	var pbModules []*schemapb.Module
	for _, m := range modules {
		deployDir := m.Config.AbsDeployDir()
		schemaPath := filepath.Join(deployDir, m.Config.Schema)
		content, err := os.ReadFile(schemaPath)
		if err != nil {
			return nil, err
		}
		pbModule := &schemapb.Module{}
		err = proto.Unmarshal(content, pbModule)
		if err != nil {
			return nil, err
		}
		pbModules = append(pbModules, pbModule)
	}
	return schema.FromProto(&schemapb.Schema{Modules: pbModules})
}

type DeclType interface {
	schema.Config | schema.Secret
}

func validateAgainstSchema(ctx context.Context, dal optional.Option[*dal.DAL], isSecret bool, ref cf.Ref, value json.RawMessage) error {
	sch, err := schemaFromDal(ctx, dal)
	if err != nil {
		sch, err = schemaFromDisk(ctx)
		if err != nil {
			return err
		}
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
	dec := json.NewDecoder(bytes.NewReader(value))
	dec.DisallowUnknownFields()
	err = dec.Decode(&v)
	if err != nil {
		return err
	}

	err = schema.ValidateJSONalue(fieldType, []string{ref.Name}, v, sch)
	if err != nil {
		return err
	}

	return nil
}
