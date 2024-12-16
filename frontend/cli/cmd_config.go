package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"

	"github.com/block/ftl/backend/admin"
	ftlv1 "github.com/block/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/block/ftl/internal/configuration"
)

type configCmd struct {
	List   configListCmd   `cmd:"" help:"List configuration."`
	Get    configGetCmd    `cmd:"" help:"Get a configuration value."`
	Set    configSetCmd    `cmd:"" help:"Set a configuration value."`
	Unset  configUnsetCmd  `cmd:"" help:"Unset a configuration value."`
	Import configImportCmd `cmd:"" help:"Import configuration values."`
	Export configExportCmd `cmd:"" help:"Export configuration values."`

	Envar  bool `help:"Print configuration as environment variables." group:"Provider:" xor:"configwriter"`
	Inline bool `help:"Write values inline in the configuration file." group:"Provider:" xor:"configwriter"`
}

func (s *configCmd) Help() string {
	return `
Configuration values are used to store non-sensitive information such as URLs,
etc.
`
}

func configRefFromRef(ref configuration.Ref) *ftlv1.ConfigRef {
	module := ref.Module.Default("")
	return &ftlv1.ConfigRef{
		Module: &module,
		Name:   ref.Name,
	}
}

func (s *configCmd) provider() optional.Option[ftlv1.ConfigProvider] {
	if s.Envar {
		return optional.Some(ftlv1.ConfigProvider_CONFIG_PROVIDER_ENVAR)
	} else if s.Inline {
		return optional.Some(ftlv1.ConfigProvider_CONFIG_PROVIDER_INLINE)
	}
	return optional.None[ftlv1.ConfigProvider]()
}

type configListCmd struct {
	Values bool   `help:"List configuration values."`
	Module string `optional:"" arg:"" placeholder:"MODULE" help:"List configuration only in this module."`
}

func (s *configListCmd) Run(ctx context.Context, adminClient admin.Client) error {
	resp, err := adminClient.ConfigList(ctx, connect.NewRequest(&ftlv1.ConfigListRequest{
		Module:        &s.Module,
		IncludeValues: &s.Values,
	}))
	if err != nil {
		return err
	}

	for _, config := range resp.Msg.Configs {
		fmt.Printf("%s", config.RefPath)
		if len(config.Value) > 0 {
			fmt.Printf(" = %s\n", config.Value)
		} else {
			fmt.Println()
		}
	}
	return nil
}

type configGetCmd struct {
	Ref configuration.Ref `arg:"" help:"Configuration reference in the form [<module>.]<name>."`
}

func (s *configGetCmd) Help() string {
	return `
Returns a JSON-encoded configuration value.
`
}

func (s *configGetCmd) Run(ctx context.Context, adminClient admin.Client) error {
	resp, err := adminClient.ConfigGet(ctx, connect.NewRequest(&ftlv1.ConfigGetRequest{
		Ref: configRefFromRef(s.Ref),
	}))
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}
	fmt.Printf("%s\n", resp.Msg.Value)
	return nil
}

type configSetCmd struct {
	JSON  bool              `help:"Assume input value is JSON. Note: For string configs, the JSON value itself must be a string (e.g., '\"hello\"' or '\"{'key': 'value'}\"')."`
	Ref   configuration.Ref `arg:"" help:"Configuration reference in the form [<module>.]<name>."`
	Value *string           `arg:"" placeholder:"VALUE" help:"Configuration value (read from stdin if omitted)." optional:""`
}

func (s *configSetCmd) Run(ctx context.Context, scmd *configCmd, adminClient admin.Client) (err error) {
	var config []byte
	if s.Value != nil {
		config = []byte(*s.Value)
	} else {
		config, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read config from stdin: %w", err)
		}
	}

	var configJSON json.RawMessage
	if s.JSON {
		var jsonValue any
		if err := json.Unmarshal(config, &jsonValue); err != nil {
			return fmt.Errorf("config is not valid JSON: %w", err)
		}
		configJSON = config
	} else {
		configJSON, err = json.Marshal(string(config))
		if err != nil {
			return fmt.Errorf("failed to encode config as JSON: %w", err)
		}
	}

	req := &ftlv1.ConfigSetRequest{
		Ref:   configRefFromRef(s.Ref),
		Value: configJSON,
	}
	if provider, ok := scmd.provider().Get(); ok {
		req.Provider = &provider
	}
	_, err = adminClient.ConfigSet(ctx, connect.NewRequest(req))
	if err != nil {
		return err
	}
	return nil
}

type configUnsetCmd struct {
	Ref configuration.Ref `arg:"" help:"Configuration reference in the form [<module>.]<name>."`
}

func (s *configUnsetCmd) Run(ctx context.Context, scmd *configCmd, adminClient admin.Client) error {
	req := &ftlv1.ConfigUnsetRequest{
		Ref: configRefFromRef(s.Ref),
	}
	if provider, ok := scmd.provider().Get(); ok {
		req.Provider = &provider
	}
	_, err := adminClient.ConfigUnset(ctx, connect.NewRequest(req))
	if err != nil {
		return err
	}
	return nil
}

type configImportCmd struct {
	Input *os.File `arg:"" placeholder:"JSON" help:"JSON to import as configuration values (read from stdin if omitted). Format: {\"<module>.<name>\": <value>, ... }" optional:"" default:"-"`
}

func (s *configImportCmd) Help() string {
	return `
Imports configuration values from a JSON object.
`
}

func (s *configImportCmd) Run(ctx context.Context, cmd *configCmd, adminClient admin.Client) error {
	input, err := io.ReadAll(s.Input)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	var entries map[string]json.RawMessage
	err = json.Unmarshal(input, &entries)
	if err != nil {
		return fmt.Errorf("could not parse JSON: %w", err)
	}
	for refPath, value := range entries {
		ref, err := configuration.ParseRef(refPath)
		if err != nil {
			return fmt.Errorf("could not parse ref %q: %w", refPath, err)
		}
		bytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("could not marshal value for %q: %w", refPath, err)
		}
		req := &ftlv1.ConfigSetRequest{
			Ref:   configRefFromRef(ref),
			Value: bytes,
		}
		if provider, ok := cmd.provider().Get(); ok {
			req.Provider = &provider
		}
		_, err = adminClient.ConfigSet(ctx, connect.NewRequest(req))
		if err != nil {
			return fmt.Errorf("could not import config for %q: %w", refPath, err)
		}
	}
	return nil
}

type configExportCmd struct {
}

func (s *configExportCmd) Help() string {
	return `
Outputs configuration values in a JSON object. A provider can be used to filter which values are included.
`
}

func (s *configExportCmd) Run(ctx context.Context, cmd *configCmd, adminClient admin.Client) error {
	req := &ftlv1.ConfigListRequest{
		IncludeValues: optional.Some(true).Ptr(),
	}
	if provider, ok := cmd.provider().Get(); ok {
		req.Provider = &provider
	}
	listResponse, err := adminClient.ConfigList(ctx, connect.NewRequest(req))
	if err != nil {
		return fmt.Errorf("could not retrieve configs: %w", err)
	}
	entries := make(map[string]json.RawMessage, 0)
	for _, config := range listResponse.Msg.Configs {
		var value json.RawMessage
		err = json.Unmarshal(config.Value, &value)
		if err != nil {
			return fmt.Errorf("could not export %q: %w", config.RefPath, err)
		}
		entries[config.RefPath] = value
	}

	output, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("could not build output: %w", err)
	}
	fmt.Println(string(output))
	return nil
}
