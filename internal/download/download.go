package download

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"connectrpc.com/connect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
)

// Artefacts downloads artefacts for a deployment from the Controller.
func Artefacts(ctx context.Context, client ftlv1connect.ControllerServiceClient, name model.DeploymentName, dest string) error {
	logger := log.FromContext(ctx)
	stream, err := client.GetDeploymentArtefacts(ctx, connect.NewRequest(&ftlv1.GetDeploymentArtefactsRequest{
		DeploymentName: name.String(),
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
