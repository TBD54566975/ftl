package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/mattn/go-isatty"
	"golang.org/x/term"

	cf "github.com/TBD54566975/ftl/common/configuration"
)

type secretCmd struct {
	List  secretListCmd  `cmd:"" help:"List secrets."`
	Get   secretGetCmd   `cmd:"" help:"Get a secret."`
	Set   secretSetCmd   `cmd:"" help:"Set a secret."`
	Unset secretUnsetCmd `cmd:"" help:"Unset a secret."`

	Envar    bool `help:"Write configuration as environment variables." group:"Provider:" xor:"secretwriter"`
	Inline   bool `help:"Write values inline in the configuration file." group:"Provider:" xor:"secretwriter"`
	Keychain bool `help:"Write to the system keychain." group:"Provider:" xor:"secretwriter"`

	//TODO: with AdminService, the following will move to the controller as "vault" and "op" will be a bool flag.
	Vault string `name:"op" help:"Store a secret in this 1Password vault. The name of the 1Password item will be the <ref> and the secret will be stored in the password field." group:"Provider:" xor:"secretwriter" placeholder:"VAULT"`
}

func (s *secretCmd) Help() string {
	return `
Secrets are used to store sensitive information such as passwords, tokens, and
keys. When setting a secret, the value is read from a password prompt if stdin
is a terminal, otherwise it is read from stdin directly. Secrets can be stored
in the project's configuration file, in the system keychain, in environment
variables, and so on.
`
}

func (s *secretCmd) providerKey() string {
	if s.Envar {
		return "envar"
	} else if s.Inline {
		return "inline"
	} else if s.Keychain {
		return "keychain"
	} else if s.Vault != "" {
		return "op"
	}
	return ""
}

type secretListCmd struct {
	Values bool   `help:"List secret values."`
	Module string `optional:"" arg:"" placeholder:"MODULE" help:"List secrets only in this module."`
}

func (s *secretListCmd) Run(ctx context.Context, scmd *secretCmd, sr cf.Resolver[cf.Secrets]) error {
	sm, err := cf.NewSecretsManager(ctx, sr, scmd.Vault)
	if err != nil {
		return err
	}
	listing, err := sm.List(ctx)
	if err != nil {
		return err
	}
	for _, secret := range listing {
		module, ok := secret.Module.Get()
		if s.Module != "" && module != s.Module {
			continue
		}
		if ok {
			fmt.Printf("%s.%s", module, secret.Name)
		} else {
			fmt.Print(secret.Name)
		}
		if s.Values {
			var value any
			err := sm.Get(ctx, secret.Ref, &value)
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

type secretGetCmd struct {
	Ref cf.Ref `arg:"" help:"Secret reference in the form [<module>.]<name>."`
}

func (s *secretGetCmd) Help() string {
	return `
Returns a JSON-encoded secret value.
`
}

func (s *secretGetCmd) Run(ctx context.Context, scmd *secretCmd, sr cf.Resolver[cf.Secrets]) error {
	sm, err := cf.NewSecretsManager(ctx, sr, scmd.Vault)
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

type secretSetCmd struct {
	JSON bool   `help:"Assume input value is JSON."`
	Ref  cf.Ref `arg:"" help:"Secret reference in the form [<module>.]<name>."`
}

func (s *secretSetCmd) Run(ctx context.Context, scmd *secretCmd, sr cf.Resolver[cf.Secrets]) error {
	sm, err := cf.NewSecretsManager(ctx, sr, scmd.Vault)
	if err != nil {
		return err
	}

	// Prompt for a secret if stdin is a terminal, otherwise read from stdin.
	var secret []byte
	if isatty.IsTerminal(0) {
		fmt.Print("Secret: ")
		secret, err = term.ReadPassword(0)
		fmt.Println()
		if err != nil {
			return err
		}
	} else {
		secret, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read secret from stdin: %w", err)
		}
	}

	var secretValue any
	if s.JSON {
		if err := json.Unmarshal(secret, &secretValue); err != nil {
			return fmt.Errorf("secret is not valid JSON: %w", err)
		}
	} else {
		secretValue = string(secret)
	}
	return sm.Set(ctx, scmd.providerKey(), s.Ref, secretValue)
}

type secretUnsetCmd struct {
	Ref cf.Ref `arg:"" help:"Secret reference in the form [<module>.]<name>."`
}

func (s *secretUnsetCmd) Run(ctx context.Context, scmd *secretCmd, sr cf.Resolver[cf.Secrets]) error {
	sm, err := cf.NewSecretsManager(ctx, sr, scmd.Vault)
	if err != nil {
		return err
	}
	return sm.Unset(ctx, scmd.providerKey(), s.Ref)
}
