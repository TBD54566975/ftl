package main

import (
	"context"
	"fmt"

	"github.com/alecthomas/errors"

	ftlv1 "github.com/TBD54566975/ftl/protos/xyz/block/ftl/v1"
)

type listCmd struct{}

func (l *listCmd) Run(ctx context.Context, client ftlv1.VerbServiceClient) error {
	resp, err := client.List(ctx, &ftlv1.ListRequest{})
	if err != nil {
		return errors.WithStack(err)
	}
	for _, verb := range resp.Verbs {
		fmt.Printf("%s.%s\n", verb.Module, verb.Name)
	}
	return nil
}
