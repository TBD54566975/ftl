package main

import (
	"context"
	"fmt"
	"time"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"
	"golang.org/x/sync/errgroup"

	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/schema"
)

type schemaCmd struct {
	Get      schemaGetCmd      `cmd:"" default:"" help:"Get the current schema from FTL."`
	Protobuf schemaProtobufCmd `cmd:"" help:"Generate protobuf schema mirroring the FTL schema structure."`
}

type schemaGetCmd struct{}

func (c *schemaGetCmd) Run(ctx context.Context, client ftlv1connect.DevelServiceClient) error {
	stream, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{}))
	if err != nil {
		return errors.WithStack(err)
	}
	wg, _ := errgroup.WithContext(ctx)
	modules := make(chan *schema.Module)
	wg.Go(func() (err error) {
		for stream.Receive() {
			resp := stream.Msg()
			module := schema.ProtoToModule(resp.Schema)
			modules <- module
			if !resp.More {
				return nil
			}
		}
		return errors.WithStack(stream.Err())
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
