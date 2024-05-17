package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/TBD54566975/ftl/internal/log"
	"io"
	"os"

	"github.com/mattn/go-isatty"
	"golang.org/x/term"

	cf "github.com/TBD54566975/ftl/common/configuration"
)

type secretCmd struct {
	cf.DefaultSecretsMixin

	List  secretListCmd  `cmd:"" help:"List secrets."`
	Get   secretGetCmd   `cmd:"" help:"Get a secret."`
	Set   secretSetCmd   `cmd:"" help:"Set a secret."`
	Unset secretUnsetCmd `cmd:"" help:"Unset a secret."`
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

type secretListCmd struct {
	Values bool   `help:"List secret values."`
	Module string `optional:"" arg:"" placeholder:"MODULE" help:"List secrets only in this module."`
}

func (s *secretListCmd) Run(ctx context.Context, scmd *secretCmd, sr cf.Resolver[cf.Secrets]) error {
	sm, err := scmd.NewSecretsManager(ctx, sr)
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
	logger := log.FromContext(ctx)

	logger.Tracef("new secrets manager sr=%v", sr)
	sm, err := scmd.NewSecretsManager(ctx, sr)
	if err != nil {
		return err
	}

	logger.Tracef("1sm = %v", sm)

	var value any

	logger.Tracef("1Getting secret %s", s.Ref)
	err = sm.Get(ctx, s.Ref, &value)
	if err != nil {
		return err
	}

	logger.Tracef("1encoding json")
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	logger.Tracef("1Encoding secret %s", s.Ref)
	err = enc.Encode(value)
	if err != nil {
		return fmt.Errorf("%s: %w", s.Ref, err)
	}

	logger.Tracef("1Returning secret %s", s.Ref)
	return nil
}

type secretSetCmd struct {
	JSON bool   `help:"Assume input value is JSON."`
	Ref  cf.Ref `arg:"" help:"Secret reference in the form [<module>.]<name>."`
}

func (s *secretSetCmd) Run(ctx context.Context, scmd *secretCmd, sr cf.Resolver[cf.Secrets]) error {
	sm, err := scmd.NewSecretsManager(ctx, sr)
	if err != nil {
		return err
	}

	if err := sm.Mutable(); err != nil {
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
	return sm.Set(ctx, s.Ref, secretValue)
}

type secretUnsetCmd struct {
	Ref cf.Ref `arg:"" help:"Secret reference in the form [<module>.]<name>."`
}

func (s *secretUnsetCmd) Run(ctx context.Context, scmd *secretCmd, sr cf.Resolver[cf.Secrets]) error {
	sm, err := scmd.NewSecretsManager(ctx, sr)
	if err != nil {
		return err
	}
	return sm.Unset(ctx, s.Ref)
}
