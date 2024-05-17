package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"connectrpc.com/connect"
	"github.com/jpillora/backoff"
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

	// otherwise, we have a match so call the verb
	resp, err := client.Call(ctx, connect.NewRequest(&ftlv1.CallRequest{
		Verb: c.Verb.ToProto(),
		Body: requestJSON,
	}))

	if cerr := new(connect.Error); errors.As(err, &cerr) && cerr.Code() == connect.CodeNotFound {
		suggestions, err := c.findSuggestions(ctx, ctlCli)

		// if we have suggestions, return a helpful error message. otherwise continue to the original error
		if err == nil {
			return fmt.Errorf("verb not found: %s\n\nDid you mean one of these?\n%s", c.Verb, suggestions)
		}
	}
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

// findSuggestions looks up the schema and finds verbs that are similar to the one that was not found
// it uses the levenshtein distance to determine similarity - if the distance is less than 40% of the length of the verb,
// it returns an error if no closely matching suggestions are found
func (c *callCmd) findSuggestions(ctx context.Context, client ftlv1connect.ControllerServiceClient) ([]string, error) {
	logger := log.FromContext(ctx)

	// lookup the verbs
	res, err := client.GetSchema(ctx, connect.NewRequest(&ftlv1.GetSchemaRequest{}))
	if err != nil {
		return nil, err
	}

	modules := res.Msg.GetSchema().GetModules()
	verbs := []string{}

	// build a list of all the verbs
	for _, module := range modules {
		for _, decl := range module.GetDecls() {
			v := decl.GetVerb()
			if v == nil {
				continue
			}

			verbName := fmt.Sprintf("%s.%s", module.Name, v.Name)
			if verbName == fmt.Sprintf("%s.%s", c.Verb.Module, c.Verb.Name) {
				break
			}

			verbs = append(verbs, module.Name+"."+v.Name)
		}
	}

	suggestions := []string{}

	logger.Debugf("Found %d verbs", len(verbs))
	needle := fmt.Sprintf("%s.%s", c.Verb.Module, c.Verb.Name)

	// only consider suggesting verbs that are within 40% of the length of the needle
	distanceThreshold := int(float64(len(needle))*0.4) + 1
	for _, verb := range verbs {
		d := levenshtein(verb, needle)
		logger.Debugf("Verb %s distance %d", verb, d)

		if d <= distanceThreshold {
			suggestions = append(suggestions, verb)
		}
	}

	if len(suggestions) > 0 {
		return suggestions, nil
	}

	return nil, fmt.Errorf("no suggestions found")
}

// Levenshtein computes the Levenshtein distance between two strings.
//
// credit goes to https://en.wikibooks.org/wiki/Algorithm_Implementation/Strings/Levenshtein_distance#Go
func levenshtein(a, b string) int {
	f := make([]int, utf8.RuneCountInString(b)+1)

	for j := range f {
		f[j] = j
	}

	for _, ca := range a {
		j := 1
		fj1 := f[0] // fj1 is the value of f[j - 1] in last iteration
		f[0]++
		for _, cb := range b {
			mn := min(f[j]+1, f[j-1]+1) // delete & insert
			if cb != ca {
				mn = min(mn, fj1+1) // change
			} else {
				mn = min(mn, fj1) // matched
			}

			fj1, f[j] = f[j], mn // save f[j] to fj1(j is about to increase), update f[j] to mn
			j++
		}
	}

	return f[len(f)-1]
}
