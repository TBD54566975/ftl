package buildengine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/runner"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/moduleconfig"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/sha256"
	"github.com/TBD54566975/ftl/internal/slices"
)

type deploymentArtefact struct {
	*ftlv1.DeploymentArtefact
	localPath string
}

type DeployClient interface {
	GetArtefactDiffs(ctx context.Context, req *connect.Request[ftlv1.GetArtefactDiffsRequest]) (*connect.Response[ftlv1.GetArtefactDiffsResponse], error)
	UploadArtefact(ctx context.Context, req *connect.Request[ftlv1.UploadArtefactRequest]) (*connect.Response[ftlv1.UploadArtefactResponse], error)
	CreateDeployment(ctx context.Context, req *connect.Request[ftlv1.CreateDeploymentRequest]) (*connect.Response[ftlv1.CreateDeploymentResponse], error)
	ReplaceDeploy(ctx context.Context, req *connect.Request[ftlv1.ReplaceDeployRequest]) (*connect.Response[ftlv1.ReplaceDeployResponse], error)
	Status(ctx context.Context, req *connect.Request[ftlv1.StatusRequest]) (*connect.Response[ftlv1.StatusResponse], error)
	UpdateDeploy(ctx context.Context, req *connect.Request[ftlv1.UpdateDeployRequest]) (*connect.Response[ftlv1.UpdateDeployResponse], error)
	GetSchema(ctx context.Context, req *connect.Request[ftlv1.GetSchemaRequest]) (*connect.Response[ftlv1.GetSchemaResponse], error)
	PullSchema(ctx context.Context, req *connect.Request[ftlv1.PullSchemaRequest]) (*connect.ServerStreamForClient[ftlv1.PullSchemaResponse], error)
	Ping(ctx context.Context, req *connect.Request[ftlv1.PingRequest]) (*connect.Response[ftlv1.PingResponse], error)
}

// Deploy a module to the FTL controller with the given number of replicas. Optionally wait for the deployment to become ready.
func Deploy(ctx context.Context, projectConfig projectconfig.Config, module Module, deploy []string, replicas int32, waitForDeployOnline bool, client DeployClient) error {
	logger := log.FromContext(ctx).Module(module.Config.Module).Scope("deploy")
	ctx = log.ContextWithLogger(ctx, logger)
	logger.Infof("Deploying module")

	moduleConfig := module.Config.Abs()
	files, err := FindFilesToDeploy(moduleConfig, deploy)
	if err != nil {
		logger.Errorf(err, "failed to find files in %s", moduleConfig)
		return err
	}

	filesByHash, err := hashFiles(moduleConfig.DeployDir, files)
	if err != nil {
		return err
	}

	gadResp, err := client.GetArtefactDiffs(ctx, connect.NewRequest(&ftlv1.GetArtefactDiffsRequest{ClientDigests: maps.Keys(filesByHash)}))
	if err != nil {
		return fmt.Errorf("failed to get artefact diffs: %w", err)
	}

	moduleSchema, err := loadProtoSchema(projectConfig, moduleConfig, replicas)
	if err != nil {
		return err
	}

	logger.Debugf("Uploading %d/%d files", len(gadResp.Msg.MissingDigests), len(files))
	for _, missing := range gadResp.Msg.MissingDigests {
		file := filesByHash[missing]
		content, err := os.ReadFile(file.localPath)
		if err != nil {
			return err
		}
		logger.Tracef("Uploading %s", relToCWD(file.localPath))
		resp, err := client.UploadArtefact(ctx, connect.NewRequest(&ftlv1.UploadArtefactRequest{
			Content: content,
		}))
		if err != nil {
			return err
		}
		logger.Debugf("Uploaded %s as %s:%s", relToCWD(file.localPath), sha256.FromBytes(resp.Msg.Digest), file.Path)
	}

	resp, err := client.CreateDeployment(ctx, connect.NewRequest(&ftlv1.CreateDeploymentRequest{
		Schema: moduleSchema,
		Artefacts: slices.Map(maps.Values(filesByHash), func(a deploymentArtefact) *ftlv1.DeploymentArtefact {
			return a.DeploymentArtefact
		}),
	}))
	if err != nil {
		return err
	}

	_, err = client.ReplaceDeploy(ctx, connect.NewRequest(&ftlv1.ReplaceDeployRequest{DeploymentKey: resp.Msg.GetDeploymentKey(), MinReplicas: replicas}))
	if err != nil {
		return err
	}

	if waitForDeployOnline {
		logger.Debugf("Waiting for deployment %s to become ready", resp.Msg.DeploymentKey)
		err = checkReadiness(ctx, client, resp.Msg.DeploymentKey, replicas, moduleSchema)
		if err != nil {
			return err
		}
		logger.Debugf("Deployment %s became ready", resp.Msg.DeploymentKey)
	}

	return nil
}

func terminateModuleDeployment(ctx context.Context, client DeployClient, module string) error {
	logger := log.FromContext(ctx).Module(module).Scope("terminate")

	status, err := client.Status(ctx, connect.NewRequest(&ftlv1.StatusRequest{}))
	if err != nil {
		return err
	}

	var key string
	for _, deployment := range status.Msg.Deployments {
		if deployment.Name == module {
			key = deployment.Key
			continue
		}
	}

	if key == "" {
		return fmt.Errorf("deployment for module %s not found: %v", module, status.Msg.Deployments)
	}

	logger.Infof("Terminating deployment %s", key)
	_, err = client.UpdateDeploy(ctx, connect.NewRequest(&ftlv1.UpdateDeployRequest{DeploymentKey: key}))
	return err
}

func loadProtoSchema(projectConfig projectconfig.Config, config moduleconfig.AbsModuleConfig, replicas int32) (*schemapb.Module, error) {
	schPath := projectConfig.SchemaPath(config.Module)
	content, err := os.ReadFile(schPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load protobuf schema from %q: %w", schPath, err)
	}
	module := &schemapb.Module{}
	err = proto.Unmarshal(content, module)
	if err != nil {
		return nil, fmt.Errorf("failed to load protobuf schema from %q: %w", schPath, err)
	}
	runtime := module.Runtime
	if runtime == nil {
		runtime = &schemapb.ModuleRuntime{}
		module.Runtime = runtime
	}
	module.Runtime = runtime
	if runtime.CreateTime == nil {
		runtime.CreateTime = timestamppb.Now()
	}
	runtime.Language = config.Language
	runtime.MinReplicas = replicas
	return module, nil
}

// FindFilesToDeploy returns a list of files to deploy for the given module.
func FindFilesToDeploy(config moduleconfig.AbsModuleConfig, deploy []string) ([]string, error) {
	var out []string
	for _, f := range deploy {
		file := filepath.Clean(filepath.Join(config.DeployDir, f))
		if !strings.HasPrefix(file, config.DeployDir) {
			return nil, fmt.Errorf("deploy path %q is not beneath deploy directory %q", file, config.DeployDir)
		}
		info, err := os.Stat(file)
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			dirFiles, err := findFilesInDir(file)
			if err != nil {
				return nil, err
			}
			out = append(out, dirFiles...)
		} else {
			out = append(out, file)
		}
	}
	return out, nil
}

func findFilesInDir(dir string) ([]string, error) {
	var out []string
	return out, filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		out = append(out, path)
		return nil
	})
}

func hashFiles(base string, files []string) (filesByHash map[string]deploymentArtefact, err error) {
	filesByHash = map[string]deploymentArtefact{}
	for _, file := range files {
		r, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer r.Close() //nolint:gosec
		hash, err := sha256.SumReader(r)
		if err != nil {
			return nil, err
		}
		info, err := r.Stat()
		if err != nil {
			return nil, err
		}
		isExecutable := info.Mode()&0111 != 0
		path, err := filepath.Rel(base, file)
		if err != nil {
			return nil, err
		}
		filesByHash[hash.String()] = deploymentArtefact{
			DeploymentArtefact: &ftlv1.DeploymentArtefact{
				Digest:     hash.String(),
				Path:       path,
				Executable: isExecutable,
			},
			localPath: file,
		}
	}
	return filesByHash, nil
}

func relToCWD(path string) string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	rel, err := filepath.Rel(cwd, path)
	if err != nil {
		return path
	}
	return rel
}

func checkReadiness(ctx context.Context, client DeployClient, deploymentKey string, replicas int32, schema *schemapb.Module) error {
	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	hasVerbs := false
	for _, dec := range schema.Decls {
		if dec.GetVerb() != nil {
			hasVerbs = true
			break
		}
	}
	for {
		select {
		case <-ticker.C:
			status, err := client.Status(ctx, connect.NewRequest(&ftlv1.StatusRequest{}))
			if err != nil {
				return err
			}

			for _, deployment := range status.Msg.Deployments {
				if deployment.Key == deploymentKey {
					if deployment.Replicas >= replicas {
						if hasVerbs {
							// Also verify the routing table is ready
							for _, route := range status.Msg.Routes {
								if route.Deployment == deploymentKey {
									return nil
								}
							}

						} else {
							return nil
						}
					}
				}
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type devModeRunner struct {
	config runner.Config
}

func (r *devModeRunner) Launch(ctx context.Context) {
	logger := log.FromContext(ctx)
	go func() {
		for restarts := 0; restarts < 10; restarts++ {
			logger.Debugf("Starting runner: %s", r.config.Key)
			err := runner.Start(ctx, r.config)
			if errors.Is(err, context.Canceled) {
				return
			}
			if err != nil {
				logger.Errorf(err, "Runner failed: %s", err)
			}
		}
		logger.Errorf(fmt.Errorf("too many restarts"), "Runner failed too many times, not restarting")
	}()
}
