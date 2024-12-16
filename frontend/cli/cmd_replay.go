package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/jpillora/backoff"

	timelinepb "github.com/block/ftl/backend/protos/xyz/block/ftl/timeline/v1"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/block/ftl/backend/timeline"
	"github.com/block/ftl/common/reflection"
	"github.com/block/ftl/common/schema"
	"github.com/block/ftl/internal/log"
	"github.com/block/ftl/internal/rpc"
	status "github.com/block/ftl/internal/terminal"
)

type replayCmd struct {
	Wait time.Duration  `short:"w" help:"Wait up to this elapsed time for the FTL cluster to become available." default:"1m"`
	Verb reflection.Ref `arg:"" required:"" help:"Full path of Verb to call." predictor:"verbs"`
}

func (c *replayCmd) Run(
	ctx context.Context,
	verbClient ftlv1connect.VerbServiceClient,
	schemaClient ftlv1connect.SchemaServiceClient,
	timelineClient *timeline.Client,
) error {
	// Wait timeout is for both pings to complete, not each ping individually
	startTime := time.Now()

	if err := rpc.Wait(ctx, backoff.Backoff{Max: time.Second * 2}, c.Wait, verbClient); err != nil {
		return fmt.Errorf("failed to wait for client: %w", err)
	}

	if err := rpc.Wait(ctx, backoff.Backoff{Max: time.Second * 2}, c.Wait-time.Since(startTime), timelineClient); err != nil {
		return fmt.Errorf("failed to wait for console service client: %w", err)
	}

	logger := log.FromContext(ctx)

	// First check the verb is valid
	// lookup the verbs
	res, err := schemaClient.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	if err != nil {
		return fmt.Errorf("failed to get schema: %w", err)
	}

	found := false
	for _, pbmodule := range res.Msg.GetSchema().GetModules() {
		module, err := schema.ModuleFromProto(pbmodule)
		if err != nil {
			logger.Errorf(err, "failed to convert module from protobuf")
			continue
		}
		if module.Name == c.Verb.Module {
			for _, v := range module.Verbs() {
				if v.Name == c.Verb.Name {
					found = true
					break
				}
			}
		}
	}
	if !found {
		suggestions, err := findSuggestions(ctx, schemaClient, c.Verb)
		// if we have suggestions, return a helpful error message. otherwise continue to the original error
		if err == nil {
			return fmt.Errorf("verb not found: %s\n\nDid you mean one of these?\n%s", c.Verb, strings.Join(suggestions, "\n"))
		}
		return fmt.Errorf("verb not found: %s", c.Verb)
	}

	events, err := timelineClient.GetTimeline(ctx, connect.NewRequest(&timelinepb.GetTimelineRequest{
		Order: timelinepb.GetTimelineRequest_ORDER_DESC,
		Filters: []*timelinepb.GetTimelineRequest_Filter{
			{
				Filter: &timelinepb.GetTimelineRequest_Filter_Call{
					Call: &timelinepb.GetTimelineRequest_CallFilter{
						DestModule: c.Verb.Module,
						DestVerb:   &c.Verb.Name,
					},
				},
			},
			{
				Filter: &timelinepb.GetTimelineRequest_Filter_EventTypes{
					EventTypes: &timelinepb.GetTimelineRequest_EventTypeFilter{
						EventTypes: []timelinepb.EventType{timelinepb.EventType_EVENT_TYPE_CALL},
					},
				},
			},
		},
	}))
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}
	if len(events.Msg.GetEvents()) == 0 {
		return fmt.Errorf("no events found for %v", c.Verb)
	}
	requestJSON := events.Msg.GetEvents()[0].GetCall().Request

	logger.Infof("Calling %s with body:", c.Verb)
	status.PrintJSON(ctx, []byte(requestJSON))
	logger.Infof("Response:")
	return callVerb(ctx, verbClient, schemaClient, c.Verb, []byte(requestJSON))
}
