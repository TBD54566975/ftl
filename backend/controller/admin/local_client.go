package admin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/proto"

	"github.com/TBD54566975/ftl/backend/controller/dal"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/buildengine"
	"github.com/TBD54566975/ftl/common/configuration"
	"github.com/TBD54566975/ftl/common/projectconfig"
)

// localClient reads and writes to local projectconfig files without making any network
// calls. It allows us to interface with local ftl-project.toml files without needing to
// start a controller.
type localClient struct {
	*AdminService
}

func newLocalClient(ctx context.Context) (*localClient, error) {
	cm := configuration.ConfigFromContext(ctx)
	sm := configuration.SecretsFromContext(ctx)
	sch, err := schemaFromDisk(ctx)
	if err != nil {
		return nil, err
	}
	return &localClient{NewAdminService(cm, sm, optional.None[*dal.DAL](), optional.Some(sch))}, nil
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
