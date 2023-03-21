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

type schemaCmd struct{}

func (c *schemaCmd) Run(client ftlv1.DevelServiceClient) error {
	ctx := context.Background()
	stream, err := client.SyncSchema(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	wg, _ := errgroup.WithContext(ctx)
	modules := make(chan schema.Module)
	wg.Go(func() (err error) {
		for {
			resp, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return errors.WithStack(err)
			}
			sc, err := schema.ParseString(resp.Module, resp.Schema)
			if err != nil {
				return errors.Wrap(err, resp.Schema)
			}
			for _, module := range sc.Modules {
				modules <- module
			}
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
