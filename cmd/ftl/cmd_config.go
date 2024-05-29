package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"connectrpc.com/connect"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/ftlv1connect"
	cf "github.com/TBD54566975/ftl/common/configuration"
)

type configCmd struct {
	cf.DefaultConfigMixin

	List  configListCmd  `cmd:"" help:"List configuration."`
	Get   configGetCmd   `cmd:"" help:"Get a configuration value."`
	Set   configSetCmd   `cmd:"" help:"Set a configuration value."`
	Unset configUnsetCmd `cmd:"" help:"Unset a configuration value."`

	Envar  bool `help:"Print configuration as environment variables." group:"Provider:" xor:"configwriter"`
	Inline bool `help:"Write values inline in the configuration file." group:"Provider:" xor:"configwriter"`
}

func (s *configCmd) Help() string {
	return `
Configuration values are used to store non-sensitive information such as URLs,
etc.
`
}

func (s *configCmd) providerKey() string {
	if s.Envar {
		return "envar"
	} else if s.Inline {
		return "inline"
	}
	return ""
}

type configListCmd struct {
	Values bool   `help:"List configuration values."`
	Module string `optional:"" arg:"" placeholder:"MODULE" help:"List configuration only in this module."`
}

func (s *configListCmd) Run(ctx context.Context, admin ftlv1connect.AdminServiceClient) error {
	resp, err := admin.ConfigList(ctx, connect.NewRequest(&ftlv1.ListConfigRequest{
		// Provider: TODO(saf)
		Module:        &s.Module,
		IncludeValues: &s.Values,
	}))
	if err != nil {
		return fmt.Errorf("failed to list config: %w", err)
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

func (s *configGetCmd) Run(ctx context.Context, admin ftlv1connect.AdminServiceClient) error {
	module := s.Ref.Module.Default("")
	resp, err := admin.ConfigGet(ctx, connect.NewRequest(&ftlv1.GetConfigRequest{
		// Provider: TODO(saf)
		Ref: &ftlv1.ConfigRef{
			Module: &module,
			Name:   s.Ref.Name,
		},
	}))
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	var value any
	err = json.Unmarshal(resp.Msg.Value, &value)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config value: %w", err)
	}
	fmt.Println(value)
	return nil
}

type configSetCmd struct {
	JSON  bool    `help:"Assume input value is JSON."`
	Ref   cf.Ref  `arg:"" help:"Configuration reference in the form [<module>.]<name>."`
	Value *string `arg:"" placeholder:"VALUE" help:"Configuration value (read from stdin if omitted)." optional:""`
}

func (s *configSetCmd) Run(ctx context.Context, scmd *configCmd, admin ftlv1connect.AdminServiceClient) error {
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

	var configValue []byte
	if s.JSON {
		if err := json.Unmarshal(config, &configValue); err != nil {
			return fmt.Errorf("config is not valid JSON: %w", err)
		}
	} else {
		configValue = config
	}

	var cp ftlv1.ConfigProvider
	if scmd.Inline {
		cp = ftlv1.ConfigProvider_CONFIG_INLINE
	} else if scmd.Envar {
		cp = ftlv1.ConfigProvider_CONFIG_ENVAR
	}

	module := s.Ref.Module.Default("")
	_, err = admin.ConfigSet(ctx, connect.NewRequest(&ftlv1.SetConfigRequest{
		Provider: &cp,
		Ref: &ftlv1.ConfigRef{
			Module: &module,
			Name:   s.Ref.Name,
		},
		Value: configValue,
	}))
	if err != nil {
		return fmt.Errorf("failed to set config: %w", err)
	}
	return nil
}

type configUnsetCmd struct {
	Ref cf.Ref `arg:"" help:"Configuration reference in the form [<module>.]<name>."`
}

func (s *configUnsetCmd) Run(ctx context.Context, scmd *configCmd, admin ftlv1connect.AdminServiceClient) error {
	module := s.Ref.Module.Default("")
	_, err := admin.ConfigUnset(ctx, connect.NewRequest(&ftlv1.UnsetConfigRequest{
		// Provider: TODO(saf)
		Ref: &ftlv1.ConfigRef{
			Module: &module,
			Name:   s.Ref.Name,
		},
	}))
	if err != nil {
		return fmt.Errorf("failed to unset config: %w", err)
	}
	return nil
}
