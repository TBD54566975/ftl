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
	configMap, err := cf.ConfigFromContext(ctx).MapForModule(ctx, module.Name)
	if err != nil {
		return nil, err
	}

	// secrets
	secretsMap, err := cf.SecretsFromContext(ctx).MapForModule(ctx, module.Name)
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
