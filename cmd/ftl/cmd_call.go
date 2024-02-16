package main

import (
	"context"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"
	"github.com/titanous/json5"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/go-runtime/ftl"
	"github.com/TBD54566975/ftl/internal/log"
)

type callCmd struct {
	Verb    ftl.VerbRef `arg:"" required:"" help:"Full path of Verb to call."`
	Request string      `arg:"" optional:"" help:"JSON5 request payload." default:"{}"`
}

func (c *callCmd) Run(ctx context.Context, client ftlv1connect.VerbServiceClient) error {
	logger := log.FromContext(ctx)
	request := map[string]any{}
	err := json5.Unmarshal([]byte(c.Request), &request)
	if err != nil {
		return fmt.Errorf("%s: %w", "invalid request", err)
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("%s: %w", "invalid request", err)
	}
	resp, err := client.Call(ctx, connect.NewRequest(&ftlv1.CallRequest{
		Verb: c.Verb.ToProto(),
		Body: requestJSON,
	}))
	if err != nil {
		return err
	}
	switch resp := resp.Msg.Response.(type) {
	case *ftlv1.CallResponse_Error_:
		if resp.Error.Stack != nil && logger.GetLevel() <= log.Debug {
			fmt.Println(*resp.Error.Stack)
		}
		return fmt.Errorf("verb error: %s", resp.Error.Message)

	case *ftlv1.CallResponse_Body:
		fmt.Println(string(resp.Body))
	}
	return nil
}
