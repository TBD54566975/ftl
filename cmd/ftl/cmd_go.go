package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/TBD54566975/ftl/common/model"
	"github.com/TBD54566975/ftl/go-runtime/compile"
	"github.com/TBD54566975/ftl/go-runtime/compile/generate"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/slices"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/schema"
)

type goCmd struct {
	Schema   goSchemaCmd   `cmd:"" help:"Extract the FTL schema from a Go module."`
	Generate goGenerateCmd `cmd:"" help:"Generate Go stubs for a module."`
	Deploy   goDeployCmd   `cmd:"" help:"Compile and deploy a Go module."`
}

type goSchemaCmd struct {
	Dir []string `arg:"" help:"Path to root directory of module." type:"existingdir"`
}

func (g *goSchemaCmd) Run() error {
	s := &schema.Schema{}
	for _, dir := range g.Dir {
		module, err := compile.ExtractModuleSchema(dir)
		if err != nil {
			return errors.WithStack(err)
		}
		s.Modules = append(s.Modules, module)
	}
	if err := schema.Validate(s); err != nil {
		return errors.WithStack(err)
	}
	fmt.Println(s)
	return nil
}

type goGenerateCmd struct {
	Schema *os.File `arg:"" required:"" help:"Path to FTL schema file." default:"-"`
}

func (g *goGenerateCmd) Run() error {
	s, err := schema.Parse("", g.Schema)
	if err != nil {
		return errors.WithStack(err)
	}
	for _, module := range s.Modules {
		if err := generate.GenerateExternalModule(os.Stdout, module); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

type goDeployCmd struct {
	MinReplicas int32 `arg:"" help:"Minimum number of replicas to deploy." default:"1"`
	compile.Config
}

func (g *goDeployCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	logger := log.FromContext(ctx)
	deployment, err := compile.Compile(ctx, g.Config)
	if err != nil {
		return errors.WithStack(err)
	}
	defer deployment.Close()

	digests := slices.Map(deployment.Artefacts, func(t *model.Artefact) string { return t.Digest.String() })
	gadResp, err := client.GetArtefactDiffs(ctx, connect.NewRequest(&ftlv1.GetArtefactDiffsRequest{ClientDigests: digests}))
	if err != nil {
		return errors.WithStack(err)
	}
	artefactsToUpload := slices.Filter(deployment.Artefacts, func(t *model.Artefact) bool {
		for _, missing := range gadResp.Msg.MissingDigests {
			if t.Digest.String() == missing {
				return true
			}
		}
		return false
	})
	for _, artefact := range artefactsToUpload {
		content, err := io.ReadAll(artefact.Content)
		if err != nil {
			return errors.Wrapf(err, "failed to read artefact %q", artefact.Path)
		}
		_, err = client.UploadArtefact(ctx, connect.NewRequest(&ftlv1.UploadArtefactRequest{Content: content}))
		if err != nil {
			return errors.Wrapf(err, "failed to upload artefact %q", artefact.Path)
		}
		logger.Infof("Uploaded %s:%s", artefact.Digest, artefact.Path)
	}
	module := deployment.Schema.ToProto().(*pschema.Module) //nolint:forcetypeassert
	module.Runtime = &pschema.ModuleRuntime{
		Language:   deployment.Language,
		CreateTime: timestamppb.Now(),
	}
	cdResp, err := client.CreateDeployment(ctx, connect.NewRequest(&ftlv1.CreateDeploymentRequest{
		Schema: module,
		Artefacts: slices.Map(deployment.Artefacts, func(t *model.Artefact) *ftlv1.DeploymentArtefact {
			return &ftlv1.DeploymentArtefact{
				Digest:     t.Digest.String(),
				Path:       t.Path,
				Executable: t.Executable,
			}
		}),
	}))
	if err != nil {
		return errors.Wrap(err, "failed to create deployment")
	}
	logger.Infof("Created deployment %s", cdResp.Msg.DeploymentKey)

	_, err = client.ReplaceDeploy(ctx, connect.NewRequest(&ftlv1.ReplaceDeployRequest{
		DeploymentKey: cdResp.Msg.DeploymentKey,
		MinReplicas:   g.MinReplicas,
	}))
	if err != nil {
		return errors.Wrapf(err, "failed to deploy %q", cdResp.Msg.DeploymentKey)
	}
	return nil
}
