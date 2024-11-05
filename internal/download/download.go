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
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/model"
	"github.com/TBD54566975/ftl/internal/sha256"
)

// Artefacts downloads artefacts for a deployment from the Controller.
func Artefacts(ctx context.Context, client ftlv1connect.ControllerServiceClient, key model.DeploymentKey, dest string, registry artefacts.RegistryConfig) error {
	logger := log.FromContext(ctx)
	response, err := client.GetDeployment(ctx, connect.NewRequest(&ftlv1.GetDeploymentRequest{
		DeploymentKey: key.String(),
	}))
	if err != nil {
		return err
	}
	start := time.Now()
	count := 0
	service := artefacts.NewOCIRegistryStorage(registry)
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
