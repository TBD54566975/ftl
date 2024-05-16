package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/jpillora/backoff"
	lev "github.com/texttheater/golang-levenshtein/levenshtein"
	"github.com/titanous/json5"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type callCmd struct {
	Wait    time.Duration  `short:"w" help:"Wait up to this elapsed time for the FTL cluster to become available." default:"1m"`
	Verb    reflection.Ref `arg:"" required:"" help:"Full path of Verb to call."`
	Request string         `arg:"" optional:"" help:"JSON5 request payload." default:"{}"`
}

func (c *callCmd) Run(ctx context.Context, client ftlv1connect.VerbServiceClient, ctlCli ftlv1connect.ControllerServiceClient) error {
	ctx, cancel := context.WithTimeout(ctx, c.Wait)
	defer cancel()
	if err := rpc.Wait(ctx, backoff.Backoff{Max: time.Second * 2}, client); err != nil {
		return err
	}

	logger := log.FromContext(ctx)
	request := map[string]any{}
	err := json5.Unmarshal([]byte(c.Request), &request)
	if err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	logger.Debugf("Calling %s", c.Verb)
	// lookup the verbs
	res, err := ctlCli.GetVerbs(ctx, connect.NewRequest(&ftlv1.VerbsRequest{}))
	if err != nil {
		return err
	}

	var foundVerb *reflection.Ref
	suggestions := []string{}
	verbs := res.Msg.GetVerbs()
	logger.Debugf("Found %d verbs", len(verbs))
	needle := []rune(fmt.Sprintf("%s.%s", c.Verb.Module, c.Verb.Name))

	// only consider suggesting verbs that are within 40% of the length of the needle
	distanceThreshold := int(float64(len(needle))*0.4) + 1
	for _, verb := range verbs {
		d := lev.DistanceForStrings([]rune(verb), needle, lev.DefaultOptions)
		logger.Debugf("Verb %s distance %d", verb, d)
		// found a match, stop searching
		if d == 0 {
			foundVerb = &c.Verb
			break
		}

		if d <= distanceThreshold {
			suggestions = append(suggestions, verb)
		}
	}

	// no match found
	if foundVerb == nil {
		return fmt.Errorf("verb not found, did you mean one of these: %v", suggestions)
	}

	resp, err := client.Call(ctx, connect.NewRequest(&ftlv1.CallRequest{
		Verb: foundVerb.ToProto(),
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
