package main

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/TBD54566975/ftl/agent"
	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/common/sha256"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	pschema "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/schema"
)

type deployCmd struct {
	Base  string   `help:"Base directory relative to files to upload."`
	Files []string `arg:"" help:"Files to upload." type:"existingfile"`
}

func (d *deployCmd) Run(ctx context.Context, client ftlv1connect.BackplaneServiceClient) error {
	logger := log.FromContext(ctx)
	base := d.Base
	if base == "" {
		base = longestCommonPathPrefix(d.Files)
	}

	// Load the TOML file.
	config, err := findAndLoadConfig(base)
	if err != nil {
		return errors.WithStack(err)
	}
	logger.Infof("Creating deployment for module %s", config.Module)

	filesByHash, err := hashFiles(d.Files)
	if err != nil {
		return errors.WithStack(err)
	}
	gadResp, err := client.GetArtefactDiffs(ctx, connect.NewRequest(&ftlv1.GetArtefactDiffsRequest{ClientDigests: maps.Keys(filesByHash)}))
	if err != nil {
		return errors.WithStack(err)
	}

	artefacts := []*ftlv1.DeploymentArtefact{}
	logger.Infof("Uploading %d files", len(gadResp.Msg.MissingDigests))
	for _, missing := range gadResp.Msg.MissingDigests {
		file := filesByHash[missing]
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return errors.WithStack(err)
		}
		path, err := filepath.Rel(base, file)
		if err != nil {
			return errors.WithStack(err)
		}
		if !filepath.IsLocal(path) {
			return errors.Errorf("path %q is not local to %q", file, base)
		}
		logger.Debugf("Uploading %s", file)
		resp, err := client.UploadArtefact(ctx, connect.NewRequest(&ftlv1.UploadArtefactRequest{
			Content: content,
		}))
		if err != nil {
			return errors.WithStack(err)
		}
		logger.Infof("Uploaded %s as %s:%s", relToCWD(file), sha256.FromBytes(resp.Msg.Digest), path)
		artefacts = append(artefacts, &ftlv1.DeploymentArtefact{
			Digest: missing,
		})
	}
	resp, err := client.CreateDeployment(ctx, connect.NewRequest(&ftlv1.CreateDeploymentRequest{
		// TODO(aat): Use real data for this.
		Schema: &pschema.Module{
			Name: config.Module,
			Runtime: &pschema.ModuleRuntime{
				CreateTime: timestamppb.Now(),
				Language:   config.Language,
			},
		},
		Artefacts: artefacts,
	}))
	if err != nil {
		return errors.WithStack(err)
	}
	logger.Infof("Created deployment %s", resp.Msg.DeploymentKey)
	return nil
}

func hashFiles(files []string) (filesByHash map[string]string, err error) {
	filesByHash = map[string]string{}
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
		filesByHash[hash.String()] = file
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

func findAndLoadConfig(root string) (agent.ModuleConfig, error) {
	for dir := root; dir != "/"; dir = filepath.Dir(dir) {
		config, err := agent.LoadConfig(dir)
		if err == nil {
			return config, nil
		}
	}
	return agent.ModuleConfig{}, errors.Errorf("no ftl.toml found in %s or any parent directory", root)
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
