package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"connectrpc.com/connect"
	"github.com/alecthomas/errors"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/TBD54566975/ftl/backend/common/log"
	"github.com/TBD54566975/ftl/backend/common/moduleconfig"
	"github.com/TBD54566975/ftl/backend/common/sha256"
	"github.com/TBD54566975/ftl/backend/common/slices"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

type deployCmd struct {
	Replicas  int32  `short:"n" help:"Number of replicas to deploy." default:"1"`
	ModuleDir string `arg:"" help:"Directory containing ftl.toml" type:"existingdir" default:"."`
}

func (d *deployCmd) Run(ctx context.Context, client ftlv1connect.ControllerServiceClient) error {
	logger := log.FromContext(ctx)

	// Load the TOML file.
	config, err := moduleconfig.LoadConfig(d.ModuleDir)
	if err != nil {
		return errors.WithStack(err)
	}
	logger.Infof("Creating deployment for module %s", config.Module)

	setConfigDefaults(&config)

	if len(config.Deploy) == 0 {
		return errors.Errorf("no deploy paths defined in config")
	}

	build := buildCmd{ModuleDir: d.ModuleDir}
	err = build.Run(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	deployDir := filepath.Join(d.ModuleDir, config.DeployDir)
	files, err := findFiles(deployDir, config.Deploy)
	if err != nil {
		return errors.WithStack(err)
	}

	filesByHash, err := hashFiles(deployDir, files)
	if err != nil {
		return errors.WithStack(err)
	}

	gadResp, err := client.GetArtefactDiffs(ctx, connect.NewRequest(&ftlv1.GetArtefactDiffsRequest{ClientDigests: maps.Keys(filesByHash)}))
	if err != nil {
		return errors.WithStack(err)
	}

	module, err := d.loadProtoSchema(deployDir, config)
	if err != nil {
		return errors.WithStack(err)
	}

	logger.Infof("Uploading %d/%d files", len(gadResp.Msg.MissingDigests), len(files))
	for _, missing := range gadResp.Msg.MissingDigests {
		file := filesByHash[missing]
		content, err := os.ReadFile(file.localPath)
		if err != nil {
			return errors.WithStack(err)
		}
		logger.Debugf("Uploading %s", relToCWD(file.localPath))
		resp, err := client.UploadArtefact(ctx, connect.NewRequest(&ftlv1.UploadArtefactRequest{
			Content: content,
		}))
		if err != nil {
			return errors.WithStack(err)
		}
		logger.Infof("Uploaded %s as %s:%s", relToCWD(file.localPath), sha256.FromBytes(resp.Msg.Digest), file.Path)
	}
	resp, err := client.CreateDeployment(ctx, connect.NewRequest(&ftlv1.CreateDeploymentRequest{
		Schema: module,
		Artefacts: slices.Map(maps.Values(filesByHash), func(a deploymentArtefact) *ftlv1.DeploymentArtefact {
			return a.DeploymentArtefact
		}),
	}))
	if err != nil {
		return errors.WithStack(err)
	}
	logger.Infof("Created deployment %s", resp.Msg.DeploymentName)
	_, err = client.ReplaceDeploy(ctx, connect.NewRequest(&ftlv1.ReplaceDeployRequest{DeploymentName: resp.Msg.GetDeploymentName(), MinReplicas: d.Replicas}))
	return errors.WithStack(err)
}

func (d *deployCmd) loadProtoSchema(deployDir string, config moduleconfig.ModuleConfig) (*schemapb.Module, error) {
	schema := filepath.Join(deployDir, config.Schema)
	content, err := os.ReadFile(schema)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	module := &schemapb.Module{}
	err = proto.Unmarshal(content, module)
	if err != nil {
		return nil, errors.WithStack(err)
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
			return nil, errors.WithStack(err)
		}
		defer r.Close() //nolint:gosec
		hash, err := sha256.SumReader(r)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		info, err := r.Stat()
		if err != nil {
			return nil, errors.WithStack(err)
		}
		isExecutable := info.Mode()&0111 != 0
		path, err := filepath.Rel(base, file)
		if err != nil {
			return nil, errors.WithStack(err)
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
			return nil, errors.WithStack(err)
		}
		if info.IsDir() {
			dirFiles, err := findFilesInDir(file)
			if err != nil {
				return nil, errors.WithStack(err)
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
	return out, errors.WithStack(filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		out = append(out, path)
		return nil
	}))
}

func setConfigDefaults(config *moduleconfig.ModuleConfig) {
	switch config.Language {
	case "kotlin":
		if config.DeployDir == "" {
			config.DeployDir = "target"
		}
		if len(config.Deploy) == 0 {
			config.Deploy = []string{"main", "classes", "dependency", "classpath.txt"}
		}
		if config.Schema == "" {
			config.Schema = "schema.pb"
		}
	case "go":
		if config.DeployDir == "" {
			config.DeployDir = "build"
		}
		if len(config.Deploy) == 0 {
			config.Deploy = []string{"main", "schema.pb"}
		}
	}
}
