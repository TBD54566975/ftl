package main

import (
	"context"
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
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/common/moduleconfig"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/sha256"
	"github.com/TBD54566975/ftl/internal/slices"
)

type deployCmd struct {
	Replicas  int32  `short:"n" help:"Number of replicas to deploy." default:"1"`
	ModuleDir string `arg:"" help:"Directory containing ftl.toml" type:"existingdir" default:"."`
	Wait      bool   `help:"Only complete the deploy command when sufficient runners have been provisioned, i.e. the deployment is online and reachable." default:"false"`
}

func (d *deployCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	logger := log.FromContext(ctx)

	// Load the TOML file.
	config, err := moduleconfig.LoadModuleConfig(d.ModuleDir)
	if err != nil {
		return err
	}

	if len(config.Deploy) == 0 {
		return fmt.Errorf("no deploy paths defined in config")
	}

	build := buildCmd{Dirs: []string{d.ModuleDir}}
	err = build.Run(ctx)
	if err != nil {
		return err
	}

	logger.Infof("Preparing module '%s' for deployment", config.Module)

	deployDir := filepath.Join(d.ModuleDir, config.DeployDir)
	files, err := findFiles(deployDir, config.Deploy)
	if err != nil {
		return err
	}

	filesByHash, err := hashFiles(deployDir, files)
	if err != nil {
		return err
	}

	gadResp, err := client.GetArtefactDiffs(ctx, connect.NewRequest(&ftlv1.GetArtefactDiffsRequest{ClientDigests: maps.Keys(filesByHash)}))
	if err != nil {
		return err
	}

	module, err := d.loadProtoSchema(deployDir, config)
	if err != nil {
		return fmt.Errorf("failed to load protobuf schema from %q: %w", config.Schema, err)
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
		Schema: module,
		Artefacts: slices.Map(maps.Values(filesByHash), func(a deploymentArtefact) *ftlv1.DeploymentArtefact {
			return a.DeploymentArtefact
		}),
	}))
	if err != nil {
		return err
	}

	_, err = client.ReplaceDeploy(ctx, connect.NewRequest(&ftlv1.ReplaceDeployRequest{DeploymentName: resp.Msg.GetDeploymentName(), MinReplicas: d.Replicas}))
	if err != nil {
		return err
	}

	if d.Wait {
		logger.Infof("Waiting for deployment %s to become ready", resp.Msg.DeploymentName)
		err = d.checkReadiness(ctx, client, resp.Msg.DeploymentName)
		if err != nil {
			return err
		}
	}

	logger.Infof("Successfully created deployment %s", resp.Msg.DeploymentName)
	return nil
}

func (d *deployCmd) checkReadiness(ctx context.Context, client ftlv1connect.ControllerServiceClient, deploymentName string) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status, err := client.Status(ctx, connect.NewRequest(&ftlv1.StatusRequest{
				AllDeployments: true,
			}))
			if err != nil {
				return err
			}

			var found bool
			for _, deployment := range status.Msg.Deployments {
				if deployment.Key == deploymentName {
					found = true
					if deployment.Replicas >= d.Replicas {
						return nil
					}
				}
			}
			if !found {
				return fmt.Errorf("deployment %s not found: %v", deploymentName, status.Msg.Deployments)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
func (d *deployCmd) loadProtoSchema(deployDir string, config moduleconfig.ModuleConfig) (*schemapb.Module, error) {
	schema := filepath.Join(deployDir, config.Schema)
	content, err := os.ReadFile(schema)
	if err != nil {
		return nil, err
	}
	module := &schemapb.Module{}
	err = proto.Unmarshal(content, module)
	if err != nil {
		return nil, err
	}
	module.Runtime = &schemapb.ModuleRuntime{
		CreateTime:  timestamppb.Now(),
		Language:    config.Language,
		MinReplicas: d.Replicas,
	}
	return module, nil
}

type deploymentArtefact struct {
	*ftlv1.DeploymentArtefact
	localPath string
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

func longestCommonPathPrefix(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	parts := strings.Split(filepath.Dir(paths[0]), "/")
	for _, path := range paths[1:] {
		parts2 := strings.Split(path, "/")
		for i := range parts {
			if i >= len(parts2) || parts[i] != parts2[i] {
				parts = parts[:i]
				break
			}
		}
	}
	return strings.Join(parts, "/")
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

func findFiles(base string, files []string) ([]string, error) {
	var out []string
	for _, file := range files {
		file = filepath.Join(base, file)
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
