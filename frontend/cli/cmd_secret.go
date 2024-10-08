package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"connectrpc.com/connect"
	"github.com/alecthomas/types/optional"
	"github.com/mattn/go-isatty"
	"golang.org/x/term"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	cf "github.com/TBD54566975/ftl/internal/configuration"
	"github.com/TBD54566975/ftl/internal/projectconfig"
)

type secretCmd struct {
	List   secretListCmd   `cmd:"" help:"List secrets."`
	Get    secretGetCmd    `cmd:"" help:"Get a secret."`
	Set    secretSetCmd    `cmd:"" help:"Set a secret."`
	Unset  secretUnsetCmd  `cmd:"" help:"Unset a secret."`
	Import secretImportCmd `cmd:"" help:"Import secrets."`
	Export secretExportCmd `cmd:"" help:"Export secrets."`

	Envar    bool `help:"Write configuration as environment variables." group:"Provider:" xor:"secretwriter"`
	Inline   bool `help:"Write values inline in the configuration file." group:"Provider:" xor:"secretwriter"`
	Keychain bool `help:"Write to the system keychain." group:"Provider:" xor:"secretwriter"`
	Op       bool `help:"Write to the controller's 1Password vault. Requires that a vault be specified to the controller. The name of the item will be the <ref> and the secret will be stored in the password field." group:"Provider:" xor:"secretwriter"`
	ASM      bool `help:"Write to AWS secrets manager." group:"Provider:" xor:"secretwriter"`
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

func (s *secretCmd) provider() optional.Option[ftlv1.SecretProvider] {
	if s.Envar {
		return optional.Some(ftlv1.SecretProvider_SECRET_ENVAR)
	} else if s.Inline {
		return optional.Some(ftlv1.SecretProvider_SECRET_INLINE)
	} else if s.Keychain {
		return optional.Some(ftlv1.SecretProvider_SECRET_KEYCHAIN)
	} else if s.Op {
		return optional.Some(ftlv1.SecretProvider_SECRET_OP)
	} else if s.ASM {
		return optional.Some(ftlv1.SecretProvider_SECRET_ASM)
	}
	return optional.None[ftlv1.SecretProvider]()
}

type secretListCmd struct {
	Values bool   `help:"List secret values."`
	Module string `optional:"" arg:"" placeholder:"MODULE" help:"List secrets only in this module."`
}

func (s *secretListCmd) Run(ctx context.Context, projConfig projectconfig.Config) error {
	ctx, adminClient, err := setUpAdminClient(ctx, projConfig)
	if err != nil {
		return err
	}
	resp, err := adminClient.SecretsList(ctx, connect.NewRequest(&ftlv1.ListSecretsRequest{
		Module:        &s.Module,
		IncludeValues: &s.Values,
	}))
	if err != nil {
		return err
	}
	for _, secret := range resp.Msg.Secrets {
		fmt.Printf("%s", secret.RefPath)
		if secret.Value != nil && len(secret.Value) > 0 {
			fmt.Printf(" = %s\n", secret.Value)
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

func (s *secretGetCmd) Run(ctx context.Context, projConfig projectconfig.Config) error {
	ctx, adminClient, err := setUpAdminClient(ctx, projConfig)
	if err != nil {
		return err
	}
	resp, err := adminClient.SecretGet(ctx, connect.NewRequest(&ftlv1.GetSecretRequest{
		Ref: configRefFromRef(s.Ref),
	}))
	if err != nil {
		return fmt.Errorf("failed to get secret: %w", err)
	}
	fmt.Printf("%s\n", resp.Msg.Value)
	return nil
}

type secretSetCmd struct {
	JSON bool   `help:"Assume input value is JSON."`
	Ref  cf.Ref `arg:"" help:"Secret reference in the form [<module>.]<name>."`
}

func (s *secretSetCmd) Run(ctx context.Context, scmd *secretCmd, projConfig projectconfig.Config) error {
	// Prompt for a secret if stdin is a terminal, otherwise read from stdin.
	ctx, adminClient, err := setUpAdminClient(ctx, projConfig)
	if err != nil {
		return err
	}
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

	var secretJSON json.RawMessage
	if s.JSON {
		var jsonValue any
		if err := json.Unmarshal(secret, &jsonValue); err != nil {
			return fmt.Errorf("secret is not valid JSON: %w", err)
		}
		secretJSON = secret
	} else {
		secretJSON, err = json.Marshal(string(secret))
		if err != nil {
			return fmt.Errorf("failed to encode secret as JSON: %w", err)
		}
	}

	req := &ftlv1.SetSecretRequest{
		Ref:   configRefFromRef(s.Ref),
		Value: secretJSON,
	}
	if provider, ok := scmd.provider().Get(); ok {
		req.Provider = &provider
	}
	_, err = adminClient.SecretSet(ctx, connect.NewRequest(req))
	if err != nil {
		return err
	}
	return nil
}

type secretUnsetCmd struct {
	Ref cf.Ref `arg:"" help:"Secret reference in the form [<module>.]<name>."`
}

func (s *secretUnsetCmd) Run(ctx context.Context, scmd *secretCmd, projConfig projectconfig.Config) error {
	ctx, adminClient, err := setUpAdminClient(ctx, projConfig)
	if err != nil {
		return err
	}
	req := &ftlv1.UnsetSecretRequest{
		Ref: configRefFromRef(s.Ref),
	}
	if provider, ok := scmd.provider().Get(); ok {
		req.Provider = &provider
	}
	_, err = adminClient.SecretUnset(ctx, connect.NewRequest(req))
	if err != nil {
		return err
	}
	return nil
}

type secretImportCmd struct {
	Input *os.File `arg:"" placeholder:"JSON" help:"JSON to import as secrets (read from stdin if omitted). Format: {\"<module>.<name>\": <secret>, ... }" optional:"" default:"-"`
}

func (s *secretImportCmd) Help() string {
	return `
Imports secrets from a JSON object.
`
}

func (s *secretImportCmd) Run(ctx context.Context, scmd *secretCmd, projConfig projectconfig.Config) error {
	ctx, adminClient, err := setUpAdminClient(ctx, projConfig)
	if err != nil {
		return err
	}
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
		ref, err := cf.ParseRef(refPath)
		if err != nil {
			return fmt.Errorf("could not parse ref %q: %w", refPath, err)
		}
		bytes, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("could not marshal value for %q: %w", refPath, err)
		}
		req := &ftlv1.SetSecretRequest{
			Ref:   configRefFromRef(ref),
			Value: bytes,
		}
		if provider, ok := scmd.provider().Get(); ok {
			req.Provider = &provider
		}
		_, err = adminClient.SecretSet(ctx, connect.NewRequest(req))
		if err != nil {
			return fmt.Errorf("could not import secret for %q: %w", refPath, err)
		}
	}
	return nil
}

type secretExportCmd struct {
}

func (s *secretExportCmd) Help() string {
	return `
Outputs secrets in a JSON object. A provider can be used to filter which secrets are included.
`
}

func (s *secretExportCmd) Run(ctx context.Context, scmd *secretCmd, projConfig projectconfig.Config) error {
	ctx, adminClient, err := setUpAdminClient(ctx, projConfig)
	if err != nil {
		return err
	}
	req := &ftlv1.ListSecretsRequest{
		IncludeValues: optional.Some(true).Ptr(),
	}
	if provider, ok := scmd.provider().Get(); ok {
		req.Provider = &provider
	}
	listResponse, err := adminClient.SecretsList(ctx, connect.NewRequest(req))
	if err != nil {
		return fmt.Errorf("could not retrieve secrets: %w", err)
	}
	entries := make(map[string]json.RawMessage, 0)
	for _, secret := range listResponse.Msg.Secrets {
		var value json.RawMessage
		err = json.Unmarshal(secret.Value, &value)
		if err != nil {
			return fmt.Errorf("could not export %q: %w", secret.RefPath, err)
		}
		entries[secret.RefPath] = value
	}

	output, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("could not build output: %w", err)
	}
	fmt.Println(string(output))
	return nil
}
