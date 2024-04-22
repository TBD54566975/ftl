package modulecontext

import (
	"context"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	cf "github.com/TBD54566975/ftl/common/configuration"
)

type ModuleContext struct {
	configManager  *cf.Manager[cf.Configuration]
	secretsManager *cf.Manager[cf.Secrets]
	dbProvider     *DBProvider
}

func (m *ModuleContext) ApplyToContext(ctx context.Context) context.Context {
	ctx = ContextWithDBProvider(ctx, m.dbProvider)
	ctx = cf.ContextWithConfig(ctx, m.configManager)
	ctx = cf.ContextWithSecrets(ctx, m.secretsManager)
	return ctx
}

func (m *ModuleContext) ToProto(ctx context.Context) (*ftlv1.ModuleContextResponse, error) {
	configList, err := m.configManager.List(ctx)
	if err != nil {
		return nil, err
	}
	configProtos := []*ftlv1.ModuleContextResponse_Config{}
	for _, entry := range configList {
		p, err := m.configEntryToConfigProto(ctx, entry)
		if err != nil {
			return nil, err
		}
		configProtos = append(configProtos, p)
	}

	secretsList, err := m.secretsManager.List(ctx)
	if err != nil {
		return nil, err
	}
	secretProtos := []*ftlv1.ModuleContextResponse_Secret{}
	for _, entry := range secretsList {
		p, err := m.configEntryToSecretProto(ctx, entry)
		if err != nil {
			return nil, err
		}
		secretProtos = append(secretProtos, p)
	}

	dsnProtos := []*ftlv1.ModuleContextResponse_DSN{}
	for name, entry := range m.dbProvider.entries {
		dsnProtos = append(dsnProtos, &ftlv1.ModuleContextResponse_DSN{
			Name: name,
			Type: ftlv1.ModuleContextResponse_DBType(entry.dbType),
			Dsn:  entry.dsn,
		})
	}

	return &ftlv1.ModuleContextResponse{
		Configs:   configProtos,
		Secrets:   secretProtos,
		Databases: dsnProtos,
	}, nil
}

func (m *ModuleContext) configEntryToConfigProto(ctx context.Context, e cf.Entry) (*ftlv1.ModuleContextResponse_Config, error) {
	data, err := m.configManager.GetData(ctx, e.Ref)
	if err != nil {
		return nil, err
	}
	return &ftlv1.ModuleContextResponse_Config{
		Ref:  m.configRefToProto(e.Ref),
		Data: data,
	}, nil
}

func (m *ModuleContext) configEntryToSecretProto(ctx context.Context, e cf.Entry) (*ftlv1.ModuleContextResponse_Secret, error) {
	data, err := m.secretsManager.GetData(ctx, e.Ref)
	if err != nil {
		return nil, err
	}
	return &ftlv1.ModuleContextResponse_Secret{
		Ref:  m.configRefToProto(e.Ref),
		Data: data,
	}, nil
}

func (m *ModuleContext) configRefToProto(r cf.Ref) *ftlv1.ModuleContextResponse_Ref {
	protoRef := &ftlv1.ModuleContextResponse_Ref{
		Name: r.Name,
	}
	if module, ok := r.Module.Get(); ok {
		protoRef.Module = &module
	}
	return protoRef
}
