package controller

import (
	"context"
	"fmt"
	"os"
	"strings"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
	cf "github.com/TBD54566975/ftl/common/configuration"
)

func moduleContextToProto(ctx context.Context, module *schema.Module) (*ftlv1.ModuleContextResponse, error) {
	// configs
	configManager := cf.ConfigFromContext(ctx)
	configMap, err := bytesMapFromConfigManager(ctx, configManager, module.Name)
	if err != nil {
		return nil, err
	}

	// secrets
	secretsManager := cf.SecretsFromContext(ctx)
	secretsMap, err := bytesMapFromConfigManager(ctx, secretsManager, module.Name)
	if err != nil {
		return nil, err
	}

	// DSNs
	dsnProtos := []*ftlv1.ModuleContextResponse_DSN{}
	for _, decl := range module.Decls {
		dbDecl, ok := decl.(*schema.Database)
		if !ok {
			continue
		}
		key := fmt.Sprintf("FTL_POSTGRES_DSN_%s_%s", strings.ToUpper(module.Name), strings.ToUpper(dbDecl.Name))
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

	return &ftlv1.ModuleContextResponse{
		Configs:   configMap,
		Secrets:   secretsMap,
		Databases: dsnProtos,
	}, nil
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
