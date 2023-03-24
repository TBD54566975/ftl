package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/alecthomas/errors"
	"golang.org/x/sync/errgroup"

	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/schema"
)

type schemaCmd struct {
	Get      schemaGetCmd      `cmd:"" default:"" help:"Get the current schema from FTL."`
	Protobuf schemaProtobufCmd `cmd:"" help:"Generate protobuf schema mirroring the FTL schema structure."`
}

type schemaGetCmd struct{}

func (c *schemaGetCmd) Run(client ftlv1.DevelServiceClient) error {
	ctx := context.Background()
	stream, err := client.SyncSchema(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	wg, _ := errgroup.WithContext(ctx)
	modules := make(chan *schema.Module)
	wg.Go(func() (err error) {
		for {
			resp, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return errors.WithStack(err)
			}
			module := schema.ProtoToModule(resp.Schema)
			modules <- module
			if !resp.More {
				return nil
			}
		}
	})

	wait := make(chan error)
	go func() { wait <- wg.Wait() }()

	for {
		select {
		case err := <-wait:
			return errors.WithStack(err)

		case m := <-modules:
			fmt.Println(m)

		case <-time.After(time.Second):
			return nil
		}
	}
}

type schemaProtobufCmd struct{}

func (c *schemaProtobufCmd) Run() error { //nolint:unparam
	fmt.Println(schema.ProtobufSchema())
	return nil
}
