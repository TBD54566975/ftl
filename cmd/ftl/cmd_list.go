package main

import (
	"context"
	"fmt"

	"github.com/alecthomas/errors"
	"github.com/bufbuild/connect-go"

	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1/ftlv1connect"
)

type listCmd struct{}

func (l *listCmd) Run(ctx context.Context, client ftlv1connect.VerbServiceClient) error {
	resp, err := client.List(ctx, connect.NewRequest(&ftlv1.ListRequest{}))
	if err != nil {
		return errors.WithStack(err)
	}
	for _, verb := range resp.Msg.Verbs {
		fmt.Printf("%s.%s\n", verb.Module, verb.Name)
	}
	return nil
}
