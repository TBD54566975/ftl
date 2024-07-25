package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"connectrpc.com/connect"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/mattn/go-isatty"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/backend/schema"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/rpc"
)

type schemaDiffCmd struct {
	OtherEndpoint url.URL `arg:"" help:"Other endpoint URL to compare against."`
	Color         bool    `help:"Enable colored output regardless of TTY."`
}

func (d *schemaDiffCmd) Run(ctx context.Context, currentURL *url.URL) error {
	other, err := schemaForURL(ctx, d.OtherEndpoint)
	if err != nil {
		return fmt.Errorf("failed to get other schema: %w", err)
	}
	current, err := schemaForURL(ctx, *currentURL)
	if err != nil {
		return fmt.Errorf("failed to get current schema: %w", err)
	}

	edits := myers.ComputeEdits(span.URIFromPath(""), other.String(), current.String())
	diff := fmt.Sprint(gotextdiff.ToUnified(d.OtherEndpoint.String(), currentURL.String(), other.String(), edits))

	color := d.Color || isatty.IsTerminal(os.Stdout.Fd())
	if color {
		err = quick.Highlight(os.Stdout, diff, "diff", "terminal256", "solarized-dark")
		if err != nil {
			return fmt.Errorf("failed to highlight diff: %w", err)
		}
	} else {
		fmt.Print(diff)
	}

	// Similar to the `diff` command, exit with 1 if there are differences.
	if diff != "" {
		os.Exit(1)
	}

	return nil
}

func schemaForURL(ctx context.Context, url url.URL) (*schema.Schema, error) {
	client := rpc.Dial(ftlv1connect.NewControllerServiceClient, url.String(), log.Error)
	resp, err := client.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{}))
	if err != nil {
		return nil, fmt.Errorf("url %s: failed to pull schema: %w", url.String(), err)
	}

	pb := &schemapb.Schema{}
	for resp.Receive() {
		msg := resp.Msg()
		pb.Modules = append(pb.Modules, msg.Schema)
		if !msg.More {
			break
		}
	}
	if resp.Err() != nil {
		return nil, fmt.Errorf("url %s: failed to receive schema: %w", url.String(), resp.Err())
	}

	s, err := schema.FromProto(pb)
	if err != nil {
		return nil, fmt.Errorf("url %s: failed to parse schema: %w", url.String(), err)
	}

	return s, nil
}
