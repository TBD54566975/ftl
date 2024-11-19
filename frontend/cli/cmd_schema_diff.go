package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"

	"connectrpc.com/connect"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/alecthomas/types/either"
	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/mattn/go-isatty"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	schemapb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/schema"
	"github.com/TBD54566975/ftl/internal/projectconfig"
	"github.com/TBD54566975/ftl/internal/schema"
	"github.com/TBD54566975/ftl/internal/terminal"
	"github.com/TBD54566975/ftl/internal/watch"
)

type schemaDiffCmd struct {
	OtherEndpoint url.URL `arg:"" help:"Other endpoint URL to compare against. If this is not specified then ftl will perform a diff against the local schema." optional:""`
	Color         bool    `help:"Enable colored output regardless of TTY."`
}

func (d *schemaDiffCmd) Run(
	ctx context.Context,
	currentURL *url.URL,
	projConfig projectconfig.Config,
	schemaClient ftlv1connect.SchemaServiceClient,
) error {
	var other *schema.Schema
	var err error
	sameModulesOnly := false
	otherEndpoint := d.OtherEndpoint.String()
	if otherEndpoint == "" {
		otherEndpoint = "Local Changes"
		sameModulesOnly = true

		other, err = localSchema(ctx, projConfig)
	} else {
		other, err = schemaForURL(ctx, schemaClient, d.OtherEndpoint)
	}
	if err != nil {
		return fmt.Errorf("failed to get other schema: %w", err)
	}
	current, err := schemaForURL(ctx, schemaClient, *currentURL)
	if err != nil {
		return fmt.Errorf("failed to get current schema: %w", err)
	}
	if sameModulesOnly {
		tempModules := current.Modules
		current.Modules = []*schema.Module{}
		moduleMap := map[string]*schema.Module{}
		for _, i := range tempModules {
			moduleMap[i.Name] = i
		}
		for _, i := range other.Modules {
			if mod, ok := moduleMap[i.Name]; ok {
				current.Modules = append(current.Modules, mod)
			}
		}
	}

	edits := myers.ComputeEdits(span.URIFromPath(""), current.String(), other.String())
	diff := fmt.Sprint(gotextdiff.ToUnified(currentURL.String(), otherEndpoint, current.String(), edits))

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
		// Unfortunately we need to close the terminal before exit to make sure the output is printed
		// This is only applicable when we explicitly call os.Exit
		terminal.FromContext(ctx).Close()
		os.Exit(1)
	}

	return nil
}

func localSchema(ctx context.Context, projectConfig projectconfig.Config) (*schema.Schema, error) {
	errs := []error{}
	modules, err := watch.DiscoverModules(ctx, projectConfig.AbsModuleDirs())
	if err != nil {
		return nil, fmt.Errorf("failed to discover modules %w", err)
	}

	moduleSchemas := make(chan either.Either[*schema.Module, error], len(modules))
	defer close(moduleSchemas)

	for _, m := range modules {
		go func() {
			module, err := schema.ModuleFromProtoFile(projectConfig.SchemaPath(m.Module))
			if err != nil {
				moduleSchemas <- either.RightOf[*schema.Module](err)
				return
			}
			moduleSchemas <- either.LeftOf[error](module)
		}()
	}
	sch := &schema.Schema{}
	for range len(modules) {
		result := <-moduleSchemas
		switch result := result.(type) {
		case either.Left[*schema.Module, error]:
			sch.Upsert(result.Get())
		case either.Right[*schema.Module, error]:
			errs = append(errs, result.Get())
		default:
			panic(fmt.Sprintf("unexpected type %T", result))

		}
	}
	// we want schema even if there are errors as long as we have some modules
	if len(sch.Modules) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("failed to read schema, possibly due to not building: %w", errors.Join(errs...))
	}
	return sch, nil
}

func schemaForURL(ctx context.Context, schemaClient ftlv1connect.SchemaServiceClient, url url.URL) (*schema.Schema, error) {
	resp, err := schemaClient.PullSchema(ctx, connect.NewRequest(&ftlv1.PullSchemaRequest{}))
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
