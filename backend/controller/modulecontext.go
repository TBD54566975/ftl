package controller

import (
	"context"
	"fmt"
	"os"
	"strings"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/common/configuration"
	cf "github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/internal/slices"
)

func moduleContextToProto(ctx context.Context, name string, schemas []*schema.Module) (*connect.Response[ftlv1.ModuleContextResponse], error) {
	schemas = slices.Filter(schemas, func(s *schema.Module) bool {
		return s.Name == name
	})
	if len(schemas) == 0 {
		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("no schema found for module %q", name))
	} else if len(schemas) > 1 {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("multiple schemas found for module %q", name))
	}

	// configs
	configManager := configuration.ConfigFromContext(ctx)
	configList, err := configManager.List(ctx)
	if err != nil {
		return nil, err
	}
	configProtos := []*ftlv1.ModuleContextResponse_Config{}
	for _, entry := range configList {
		data, err := configManager.GetData(ctx, entry.Ref)
		if err != nil {
			return nil, err
		}
		configProtos = append(configProtos, &ftlv1.ModuleContextResponse_Config{
			Ref:  configRefToProto(entry.Ref),
			Data: data,
		})
	}

	// secrets
	secretsManager := configuration.SecretsFromContext(ctx)
	secretsList, err := secretsManager.List(ctx)
	if err != nil {
		return nil, err
	}
	secretProtos := []*ftlv1.ModuleContextResponse_Secret{}
	for _, entry := range secretsList {
		data, err := secretsManager.GetData(ctx, entry.Ref)
		if err != nil {
			return nil, err
		}
		secretProtos = append(secretProtos, &ftlv1.ModuleContextResponse_Secret{
			Ref:  configRefToProto(entry.Ref),
			Data: data,
		})
	}

	// DSNs
	dsnProtos := []*ftlv1.ModuleContextResponse_DSN{}
	for _, decl := range schemas[0].Decls {
		dbDecl, ok := decl.(*schema.Database)
		if !ok {
			continue
		}
		key := fmt.Sprintf("FTL_POSTGRES_DSN_%s_%s", strings.ToUpper(name), strings.ToUpper(dbDecl.Name))
		dsn, ok := os.LookupEnv(key)
		if !ok {
			return nil, fmt.Errorf("missing environment variable %q", key)
		}
		dsnProtos = append(dsnProtos, &ftlv1.ModuleContextResponse_DSN{
			Name: dbDecl.Name,
			Type: ftlv1.ModuleContextResponse_POSTGRES,
			Dsn:  dsn,
		})
	}

	return connect.NewResponse(&ftlv1.ModuleContextResponse{
		Configs:   configProtos,
		Secrets:   secretProtos,
		Databases: dsnProtos,
	}), nil
}

func configRefToProto(r cf.Ref) *ftlv1.ModuleContextResponse_Ref {
	protoRef := &ftlv1.ModuleContextResponse_Ref{
		Name: r.Name,
	}
	if module, ok := r.Module.Get(); ok {
		protoRef.Module = &module
	}
	return protoRef
}
