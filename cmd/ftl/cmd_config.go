package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

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

func (s *configListCmd) Run(ctx context.Context, scmd *configCmd, cr cf.Resolver[cf.Configuration]) error {
	sm, err := cf.NewConfigurationManager(ctx, cr)
	if err != nil {
		return err
	}
	listing, err := sm.List(ctx)
	if err != nil {
		return err
	}
	for _, config := range listing {
		module, ok := config.Module.Get()
		if s.Module != "" && module != s.Module {
			continue
		}
		if ok {
			fmt.Printf("%s.%s", module, config.Name)
		} else {
			fmt.Print(config.Name)
		}
		if s.Values {
			var value any
			err := sm.Get(ctx, config.Ref, &value)
			if err != nil {
				fmt.Printf(" (error: %s)\n", err)
			} else {
				data, _ := json.Marshal(value)
				fmt.Printf(" = %s\n", data)
			}
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

func (s *configGetCmd) Run(ctx context.Context, scmd *configCmd, cr cf.Resolver[cf.Configuration]) error {
	sm, err := cf.NewConfigurationManager(ctx, cr)
	if err != nil {
		return err
	}
	var value any
	err = sm.Get(ctx, s.Ref, &value)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	err = enc.Encode(value)
	if err != nil {
		return fmt.Errorf("%s: %w", s.Ref, err)
	}
	return nil
}

type configSetCmd struct {
	JSON  bool    `help:"Assume input value is JSON."`
	Ref   cf.Ref  `arg:"" help:"Configuration reference in the form [<module>.]<name>."`
	Value *string `arg:"" placeholder:"VALUE" help:"Configuration value (read from stdin if omitted)." optional:""`
}

func (s *configSetCmd) Run(ctx context.Context, scmd *configCmd, cr cf.Resolver[cf.Configuration]) error {
	sm, err := cf.NewConfigurationManager(ctx, cr)
	if err != nil {
		return err
	}

	var config []byte
	if s.Value != nil {
		config = []byte(*s.Value)
	} else {
		config, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read config from stdin: %w", err)
		}
	}

	var configValue any
	if s.JSON {
		if err := json.Unmarshal(config, &configValue); err != nil {
			return fmt.Errorf("config is not valid JSON: %w", err)
		}
	} else {
		configValue = string(config)
	}
	return sm.Set(ctx, scmd.providerKey(), s.Ref, configValue)
}

type configUnsetCmd struct {
	Ref cf.Ref `arg:"" help:"Configuration reference in the form [<module>.]<name>."`
}

func (s *configUnsetCmd) Run(ctx context.Context, scmd *configCmd, cr cf.Resolver[cf.Configuration]) error {
	sm, err := cf.NewConfigurationManager(ctx, cr)
	if err != nil {
		return err
	}

	return sm.Unset(ctx, scmd.providerKey(), s.Ref)
}
