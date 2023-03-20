package main

import (
	"context"
	"fmt"

	"github.com/alecthomas/errors"

	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
)

type schemaCmd struct{}

func (c *schemaCmd) Run(client ftlv1.DevelServiceClient) error {
	resp, err := client.Schema(context.Background(), &ftlv1.SchemaRequest{})
	if err != nil {
		return errors.WithStack(err)
	}
	fmt.Println(resp.Schema)
	return nil
}
