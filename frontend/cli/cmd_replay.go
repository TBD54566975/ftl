package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/jpillora/backoff"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	pbconsole "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/console"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/console/pbconsoleconnect"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/go-runtime/ftl/reflection"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type replayCmd struct {
	Wait time.Duration  `short:"w" help:"Wait up to this elapsed time for the FTL cluster to become available." default:"1m"`
	Verb reflection.Ref `arg:"" required:"" help:"Full path of Verb to call."`
}

func (c *replayCmd) Run(ctx context.Context, client ftlv1connect.VerbServiceClient, ctlCli ftlv1connect.ControllerServiceClient) error {
	ctx, cancel := context.WithTimeout(ctx, c.Wait)
	defer cancel()
	if err := rpc.Wait(ctx, backoff.Backoff{Max: time.Second * 2}, client); err != nil {
		return fmt.Errorf("failed to wait for client: %w", err)
	}

	consoleServiceClient := rpc.Dial(pbconsoleconnect.NewConsoleServiceClient, cli.Endpoint.String(), log.Error)
	if err := rpc.Wait(ctx, backoff.Backoff{Max: time.Second * 2}, consoleServiceClient); err != nil {
		return fmt.Errorf("failed to wait for console service client: %w", err)
	}

	logger := log.FromContext(ctx)

	// First check the verb is valid
	// lookup the verbs
	res, err := ctlCli.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
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
		suggestions, err := findSuggestions(ctx, ctlCli, c.Verb)
		// if we have suggestions, return a helpful error message. otherwise continue to the original error
		if err == nil {
			return fmt.Errorf("verb not found: %s\n\nDid you mean one of these?\n%s", c.Verb, strings.Join(suggestions, "\n"))
		}
		return fmt.Errorf("verb not found: %s", c.Verb)
	}

	events, err := consoleServiceClient.GetEvents(ctx, connect.NewRequest(&pbconsole.EventsQuery{
		Order: pbconsole.EventsQuery_DESC,
		Limit: 1,
		Filters: []*pbconsole.EventsQuery_Filter{
			{
				Filter: &pbconsole.EventsQuery_Filter_Call{
					Call: &pbconsole.EventsQuery_CallFilter{
						DestModule: c.Verb.Module,
						DestVerb:   &c.Verb.Name,
					},
				},
			},
			{
				Filter: &pbconsole.EventsQuery_Filter_EventTypes{
					EventTypes: &pbconsole.EventsQuery_EventTypeFilter{EventTypes: []pbconsole.EventType{pbconsole.EventType_EVENT_TYPE_CALL}},
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

	logger.Infof("Calling %s with body:\n%s", c.Verb, requestJSON)
	return callVerb(ctx, client, ctlCli, c.Verb, []byte(requestJSON))
}
