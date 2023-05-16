package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"

	"github.com/TBD54566975/ftl/common/log"
	"github.com/TBD54566975/ftl/common/sha256"
	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type downloadCmd struct {
	Dest       string    `short:"d" help:"Destination directory." default:"."`
	Deployment uuid.UUID `help:"Deployment to download." arg:""`
}

func (d *downloadCmd) Run(ctx context.Context, client ftlv1connect.BackplaneServiceClient) error {
	logger := log.FromContext(ctx)

	haveDigests := mapset.NewSet[string]()

	err := filepath.Walk(d.Dest, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		sum, err := sha256.SumFile(path)
		if err != nil {
			return errors.WithStack(err)
		}
		haveDigests.Add(sum.String())
		return nil
	})
	if err != nil {
		return errors.WithStack(err)
	}

	stream, err := client.GetDeploymentArtefacts(ctx, connect.NewRequest(&ftlv1.GetDeploymentArtefactsRequest{
		DeploymentKey: d.Deployment.String(),
		HaveDigests:   haveDigests.ToSlice(),
	}))
	if err != nil {
		return errors.WithStack(err)
	}
	var digest string
	var w *os.File
	for stream.Receive() {
		msg := stream.Msg()
		artefact := msg.Artefact
		if digest != artefact.Digest {
			if w != nil {
				w.Close()
			}
			if !filepath.IsLocal(artefact.Path) {
				return errors.Errorf("path %q is not local", artefact.Path)
			}
			logger.Infof("Downloading %s", filepath.Join(d.Dest, artefact.Path))
			err = os.MkdirAll(filepath.Join(d.Dest, filepath.Dir(artefact.Path)), 0700)
			if err != nil {
				return errors.WithStack(err)
			}
			var mode os.FileMode = 0600
			if artefact.Executable {
				mode = 0700
			}
			w, err = os.OpenFile(filepath.Join(d.Dest, artefact.Path), os.O_CREATE|os.O_WRONLY, mode)
			if err != nil {
				return errors.WithStack(err)
			}
			digest = artefact.Digest
		}

		if _, err := w.Write(msg.Chunk); err != nil {
			_ = w.Close()
			return errors.WithStack(err)
		}
	}
	if w != nil {
		w.Close()
	}
	return errors.WithStack(stream.Err())
}
