package buildengine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/common/moduleconfig"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/sha256"
	"github.com/TBD54566975/ftl/internal/slices"

	goslices "slices"
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
}

// Deploy a module to the FTL controller with the given number of replicas. Optionally wait for the deployment to become ready.
func Deploy(ctx context.Context, module Module, replicas int32, waitForDeployOnline bool, client DeployClient) error {
	logger := log.FromContext(ctx).Scope(module.Config.Module)
	ctx = log.ContextWithLogger(ctx, logger)
	logger.Infof("Deploying module")

	moduleConfig := module.Config.Abs()
	files, err := FindFilesToDeploy(moduleConfig)
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
		return err
	}

	moduleSchema, err := loadProtoSchema(moduleConfig, replicas)
	if err != nil {
		return fmt.Errorf("failed to load protobuf schema from %q: %w", module.Config.Schema, err)
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
			// There is a possible race here, another deployment may have uploaded it
			// double check it has not been added
			newDiffs, diffErr := client.GetArtefactDiffs(ctx, connect.NewRequest(&ftlv1.GetArtefactDiffsRequest{ClientDigests: maps.Keys(filesByHash)}))
			if diffErr != nil {
				return fmt.Errorf("failed to get artefact diffs: %w after upload failure %w", diffErr, err)
			}
			if goslices.Contains(newDiffs.Msg.MissingDigests, missing) {
				// It is still missing, return the error
				return fmt.Errorf("failed to upload artifacts %w", err)
			} else {
				logger.Debugf("Upload %s of was cancelled as another deployment uploaded it", relToCWD(file.localPath))
				continue
			}
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
		err = checkReadiness(ctx, client, resp.Msg.DeploymentKey, replicas)
		if err != nil {
			return err
		}
		logger.Debugf("Deployment %s became ready", resp.Msg.DeploymentKey)
	}

	return nil
}

func terminateModuleDeployment(ctx context.Context, client ftlv1connect.ControllerServiceClient, module string) error {
	logger := log.FromContext(ctx).Scope(module)

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

func loadProtoSchema(config moduleconfig.AbsModuleConfig, replicas int32) (*schemapb.Module, error) {
	content, err := os.ReadFile(config.Schema)
	if err != nil {
		return nil, err
	}
	module := &schemapb.Module{}
	err = proto.Unmarshal(content, module)
	if err != nil {
		return nil, err
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
func FindFilesToDeploy(moduleConfig moduleconfig.AbsModuleConfig) ([]string, error) {
	var out []string
	for _, file := range moduleConfig.Deploy {
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

func checkReadiness(ctx context.Context, client DeployClient, deploymentKey string, replicas int32) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

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
						return nil
					}
				}
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
