package download

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"connectrpc.com/connect"

	"github.com/TBD54566975/ftl/backend/controller/artefacts"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/common/sha256"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
)

// Artefacts downloads artefacts for a deployment from the Controller.
func Artefacts(ctx context.Context, client ftlv1connect.ControllerServiceClient, key model.DeploymentKey, dest string) error {
	logger := log.FromContext(ctx)
	stream, err := client.GetDeploymentArtefacts(ctx, connect.NewRequest(&ftlv1.GetDeploymentArtefactsRequest{
		DeploymentKey: key.String(),
	}))
	if err != nil {
		return err
	}
	start := time.Now()
	count := 0
	var digest string
	var w *os.File
	for stream.Receive() {
		msg := stream.Msg()
		artefact := msg.Artefact
		if digest != artefact.Digest {
			if w != nil {
				w.Close()
			}
			count++
			if !filepath.IsLocal(artefact.Path) {
				return fmt.Errorf("path %q is not local", artefact.Path)
			}
			logger.Debugf("Downloading %s", filepath.Join(dest, artefact.Path))
			err = os.MkdirAll(filepath.Join(dest, filepath.Dir(artefact.Path)), 0700)
			if err != nil {
				return err
			}
			var mode os.FileMode = 0600
			if artefact.Executable {
				mode = 0700
			}
			w, err = os.OpenFile(filepath.Join(dest, artefact.Path), os.O_CREATE|os.O_WRONLY, mode)
			if err != nil {
				return err
			}
			digest = artefact.Digest
		}

		if _, err := w.Write(msg.Chunk); err != nil {
			_ = w.Close()
			return err
		}
	}
	if w != nil {
		w.Close()
	}
	logger.Debugf("Downloaded %d artefacts in %s", count, time.Since(start))
	return stream.Err()
}

// ArtefactsFromOCI downloads artefacts for a deployment from an OCI registry.
func ArtefactsFromOCI(ctx context.Context, client ftlv1connect.ControllerServiceClient, key model.DeploymentKey, dest string, service *artefacts.OCIArtefactService) error {
	logger := log.FromContext(ctx)
	response, err := client.GetDeployment(ctx, connect.NewRequest(&ftlv1.GetDeploymentRequest{
		DeploymentKey: key.String(),
	}))
	if err != nil {
		return fmt.Errorf("failed to get deployment %q: %w", key, err)
	}
	start := time.Now()
	count := 0
	for _, artefact := range response.Msg.Artefacts {
		parseSHA256, err := sha256.ParseSHA256(artefact.Digest)
		if err != nil {
			return fmt.Errorf("failed to parse SHA256 %q: %w", artefact.Digest, err)
		}
		res, err := service.Download(ctx, parseSHA256)
		if err != nil {
			return fmt.Errorf("failed to download artifact %q: %w", artefact.Digest, err)
		}
		count++
		if !filepath.IsLocal(artefact.Path) {
			return fmt.Errorf("path %q is not local", artefact.Path)
		}
		logger.Debugf("Downloading %s", filepath.Join(dest, artefact.Path))
		err = os.MkdirAll(filepath.Join(dest, filepath.Dir(artefact.Path)), 0700)
		if err != nil {
			return fmt.Errorf("failed to download artifact %q: %w", artefact.Digest, err)
		}
		var mode os.FileMode = 0600
		if artefact.Executable {
			mode = 0700
		}
		w, err := os.OpenFile(filepath.Join(dest, artefact.Path), os.O_CREATE|os.O_WRONLY, mode)
		if err != nil {
			return fmt.Errorf("failed to download artifact %q: %w", artefact.Digest, err)
		}
		defer w.Close()
		buf := make([]byte, 1024)
		read := 0
		for {
			read, err = res.Read(buf)
			if read > 0 {
				_, e2 := w.Write(buf[:read])
				if e2 != nil {
					return fmt.Errorf("failed to download artifact %q: %w", artefact.Digest, err)
				}
			}
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to download artifact %q: %w", artefact.Digest, err)
			}
		}

	}
	logger.Debugf("Downloaded %d artefacts in %s", count, time.Since(start))
	return nil
}
