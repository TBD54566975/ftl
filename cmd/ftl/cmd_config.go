package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"connectrpc.com/connect"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/backend/controller/admin"
	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	cf "github.com/TBD54566975/ftl/common/configuration"
)

type configCmd struct {
	List  configListCmd  `cmd:"" help:"List configuration."`
	Get   configGetCmd   `cmd:"" help:"Get a configuration value."`
	Set   configSetCmd   `cmd:"" help:"Set a configuration value."`
	Unset configUnsetCmd `cmd:"" help:"Unset a configuration value."`

	Envar  bool `help:"Print configuration as environment variables." group:"Provider:" xor:"configwriter"`
	Inline bool `help:"Write values inline in the configuration file." group:"Provider:" xor:"configwriter"`
	DB     bool `help:"Write values to the database." group:"Provider:" xor:"configwriter"`
}

func (s *configCmd) Help() string {
	return `
Configuration values are used to store non-sensitive information such as URLs,
etc.
`
}

func configRefFromRef(ref cf.Ref) *ftlv1.ConfigRef {
	module := ref.Module.Default("")
	return &ftlv1.ConfigRef{
		Module: &module,
		Name:   ref.Name,
	}
}

func (s *configCmd) provider() optional.Option[ftlv1.ConfigProvider] {
	if s.Envar {
		return optional.Some(ftlv1.ConfigProvider_CONFIG_ENVAR)
	} else if s.Inline {
		return optional.Some(ftlv1.ConfigProvider_CONFIG_INLINE)
	} else if s.DB {
		return optional.Some(ftlv1.ConfigProvider_CONFIG_DB)
	}
	return optional.None[ftlv1.ConfigProvider]()
}

type configListCmd struct {
	Values bool   `help:"List configuration values."`
	Module string `optional:"" arg:"" placeholder:"MODULE" help:"List configuration only in this module."`
}

func (s *configListCmd) Run(ctx context.Context, adminClient admin.Client) error {
	resp, err := adminClient.ConfigList(ctx, connect.NewRequest(&ftlv1.ListConfigRequest{
		Module:        &s.Module,
		IncludeValues: &s.Values,
	}))
	if err != nil {
		return err
	}

	for _, config := range resp.Msg.Configs {
		fmt.Printf("%s", config.RefPath)
		if config.Value != nil && len(config.Value) > 0 {
			fmt.Printf(" = %s\n", config.Value)
		} else {
			fmt.Println()
		}
	}
	return nil
}

type configGetCmd struct {
	Ref cf.Ref `arg:"" help:"Configuration reference in the form [<module>.]<name>."`
}

func (s *configGetCmd) Help() string {
	return `
Returns a JSON-encoded configuration value.
`
}

func (s *configGetCmd) Run(ctx context.Context, adminClient admin.Client) error {
	resp, err := adminClient.ConfigGet(ctx, connect.NewRequest(&ftlv1.GetConfigRequest{
		Ref: configRefFromRef(s.Ref),
	}))
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}
	fmt.Printf("%s\n", resp.Msg.Value)
	return nil
}

type configSetCmd struct {
	JSON  bool    `help:"Assume input value is JSON."`
	Ref   cf.Ref  `arg:"" help:"Configuration reference in the form [<module>.]<name>."`
	Value *string `arg:"" placeholder:"VALUE" help:"Configuration value (read from stdin if omitted)." optional:""`
}

func (s *configSetCmd) Run(ctx context.Context, scmd *configCmd, adminClient admin.Client) error {
	var err error
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

	req := &ftlv1.SetConfigRequest{
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
	Ref cf.Ref `arg:"" help:"Configuration reference in the form [<module>.]<name>."`
}

func (s *configUnsetCmd) Run(ctx context.Context, scmd *configCmd, adminClient admin.Client) error {
	req := &ftlv1.UnsetConfigRequest{
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
