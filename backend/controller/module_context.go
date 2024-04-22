package controller

import (
	"context"
	"fmt"
	"os"
	"strings"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
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
	configManager := cf.ConfigFromContext(ctx)
	configMap, err := bytesMapFromConfigManager(ctx, configManager, name)
	if err != nil {
		return nil, err
	}

	// secrets
	secretsManager := cf.SecretsFromContext(ctx)
	secretsMap, err := bytesMapFromConfigManager(ctx, secretsManager, name)
	if err != nil {
		return nil, err
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
		Configs:   configMap,
		Secrets:   secretsMap,
		Databases: dsnProtos,
	}), nil
}

func bytesMapFromConfigManager[R cf.Role](ctx context.Context, manager *cf.Manager[R], moduleName string) (map[string][]byte, error) {
	configList, err := manager.List(ctx)
	if err != nil {
		return nil, err
	}

	// module specific values must override global values
	// put module specific values into moduleConfigMap, then merge with configMap
	configMap := map[string][]byte{}
	moduleConfigMap := map[string][]byte{}

	for _, entry := range configList {
		refModule, isModuleSpecific := entry.Module.Get()
		if isModuleSpecific && refModule != moduleName {
			continue
		}
		data, err := manager.GetData(ctx, entry.Ref)
		if err != nil {
			return nil, err
		}
		if !isModuleSpecific {
			configMap[entry.Ref.Name] = data
		} else {
			moduleConfigMap[entry.Ref.Name] = data
		}
	}

	for name, data := range moduleConfigMap {
		configMap[name] = data
	}
	return configMap, nil
}
