package buildengine

import (
	"context"
	"os"
	"testing"

	"connectrpc.com/connect"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/sha256"
	"github.com/alecthomas/assert/v2"
)

type mockDeployClient struct {
	MissingDigests []string
	DeploymentKey  string
}

func (m *mockDeployClient) GetArtefactDiffs(context.Context, *connect.Request[ftlv1.GetArtefactDiffsRequest]) (*connect.Response[ftlv1.GetArtefactDiffsResponse], error) {
	return connect.NewResponse(&ftlv1.GetArtefactDiffsResponse{
		MissingDigests: m.MissingDigests,
	}), nil
}

func (m *mockDeployClient) UploadArtefact(ctx context.Context, req *connect.Request[ftlv1.UploadArtefactRequest]) (*connect.Response[ftlv1.UploadArtefactResponse], error) {
	sha256digest := sha256.Sum(req.Msg.Content)
	return connect.NewResponse(&ftlv1.UploadArtefactResponse{Digest: sha256digest[:]}), nil
}

func (m *mockDeployClient) CreateDeployment(context.Context, *connect.Request[ftlv1.CreateDeploymentRequest]) (*connect.Response[ftlv1.CreateDeploymentResponse], error) {
	return connect.NewResponse(&ftlv1.CreateDeploymentResponse{DeploymentKey: m.DeploymentKey}), nil
}

func (m *mockDeployClient) ReplaceDeploy(context.Context, *connect.Request[ftlv1.ReplaceDeployRequest]) (*connect.Response[ftlv1.ReplaceDeployResponse], error) {
	return nil, nil
}

func (m *mockDeployClient) Status(context.Context, *connect.Request[ftlv1.StatusRequest]) (*connect.Response[ftlv1.StatusResponse], error) {
	resp := &ftlv1.StatusResponse{
		Deployments: []*ftlv1.StatusResponse_Deployment{
			{Key: m.DeploymentKey, Replicas: 1},
		},
	}
	return connect.NewResponse(resp), nil
}

func TestDeploy(t *testing.T) {
	sch := &schema.Schema{
		Modules: []*schema.Module{
			schema.Builtins(),
			{Name: "another", Decls: []schema.Decl{
				&schema.Data{Name: "EchoRequest"},
				&schema.Data{Name: "EchoResponse"},
				&schema.Verb{
					Name:     "echo",
					Request:  &schema.DataRef{Name: "EchoRequest"},
					Response: &schema.DataRef{Name: "EchoResponse"},
				},
			}},
		},
	}
	ctx := log.ContextWithLogger(context.Background(), log.Configure(os.Stderr, log.Config{}))

	modulePath := "testdata/modules/another"
	module, err := LoadModule(ctx, modulePath)
	assert.NoError(t, err)

	// Build first to make sure the files are there.
	err = Build(ctx, sch, module)
	assert.NoError(t, err)

	sum, err := sha256.SumFile(modulePath + "/_ftl/main")
	assert.NoError(t, err)

	client := &mockDeployClient{
		MissingDigests: []string{sum.String()},
		DeploymentKey:  "test-deployment",
	}

	err = Deploy(ctx, module, int32(1), true, client)
	assert.NoError(t, err)
}
