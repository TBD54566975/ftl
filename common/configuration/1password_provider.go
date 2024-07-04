package configuration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/kballard/go-shellquote"

	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/slices"
)

// OnePasswordProvider is a configuration provider that reads passwords from
// 1Password vaults via the "op" command line tool.
type OnePasswordProvider struct {
	Vault       string
	ProjectName string
}

func (OnePasswordProvider) Role() Secrets { return Secrets{} }
func (o OnePasswordProvider) Key() string { return "op" }
func (o OnePasswordProvider) Delete(ctx context.Context, ref Ref) error {
	return nil
}

func (o OnePasswordProvider) itemName() string {
	return o.ProjectName + ".secrets"
}

// Load returns the secret stored in 1password.
func (o OnePasswordProvider) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	if err := checkOpBinary(); err != nil {
		return nil, err
	}

	vault := key.Host
	full, err := o.getItem(ctx, vault)
	if err != nil {
		return nil, fmt.Errorf("get item failed: %w", err)
	}

	secret, ok := full.value(ref)
	if !ok {
		return nil, fmt.Errorf("field %q not found in 1Password item %q: %v", ref, o.itemName(), full.Fields)
	}

	// Just to verify that it is JSON encoded.
	var decoded interface{}
	err = json.Unmarshal(secret, &decoded)
	if err != nil {
		return nil, fmt.Errorf("secret is not JSON encoded: %w", err)
	}

	return secret, nil
}

var vaultRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-.]+$`)

// Store will save the given secret in 1Password via the `op` command.
//
// op does not support "create or update" as a single command. Neither does it support specifying an ID on create.
// Because of this, we need check if the item exists before creating it, and update it if it does.
func (o OnePasswordProvider) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	if err := checkOpBinary(); err != nil {
		return nil, err
	}
	if o.Vault == "" {
		return nil, fmt.Errorf("vault missing, specify vault as a flag to the controller")
	}
	if !vaultRegex.MatchString(o.Vault) {
		return nil, fmt.Errorf("vault name %q contains invalid characters. a-z A-Z 0-9 _ . - are valid", o.Vault)
	}

	url := &url.URL{Scheme: "op", Host: o.Vault}

	// make sure item exists
	_, err := o.getItem(ctx, o.Vault)
	if errors.As(err, new(itemNotFoundError)) {
		err = o.createItem(ctx, o.Vault)
		if err != nil {
			return nil, fmt.Errorf("create item failed: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("get item failed: %w", err)
	}

	err = o.storeSecret(ctx, o.Vault, ref, value)
	if err != nil {
		return nil, fmt.Errorf("edit item failed: %w", err)
	}

	return url, nil
}

func checkOpBinary() error {
	_, err := exec.LookPath("op")
	if err != nil {
		return fmt.Errorf("1Password CLI tool \"op\" not found: %w", err)
	}
	return nil
}

type itemNotFoundError struct {
	vault string
	name  string
}

func (e itemNotFoundError) Error() string {
	return fmt.Sprintf("item %q not found in vault %q", e.name, e.vault)
}

// item is the JSON response from `op item get`.
type item struct {
	Fields []entry `json:"fields"`
}

type entry struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

func (i item) value(ref Ref) ([]byte, bool) {
	secret, ok := slices.Find(i.Fields, func(item entry) bool {
		return item.Label == ref.String()
	})
	return []byte(secret.Value), ok
}

// getItem gets the single 1Password item for all project secrets
// op --format json item get --vault Personal "ftl.projectname.secrets"
func (o OnePasswordProvider) getItem(ctx context.Context, vault string) (*item, error) {
	logger := log.FromContext(ctx)
	args := []string{
		"item", "get", o.itemName(),
		"--vault", vault,
		"--format", "json",
	}
	output, err := exec.Capture(ctx, ".", "op", args...)
	logger.Debugf("Getting item with args %s", shellquote.Join(args...))
	if err != nil {
		// This is specifically not itemNotFoundError, to distinguish between vault not found and item not found.
		if strings.Contains(string(output), "isn't a vault") {
			return nil, fmt.Errorf("vault %q not found: %w", vault, err)
		}

		// Item not found, seen two ways of reporting this:
		if strings.Contains(string(output), "not found in vault") {
			return nil, itemNotFoundError{vault, o.itemName()}
		}
		if strings.Contains(string(output), "isn't an item") {
			return nil, itemNotFoundError{vault, o.itemName()}
		}

		return nil, fmt.Errorf("run `op` with args %s: %w", shellquote.Join(args...), err)
	}

	var full item
	if err := json.Unmarshal(output, &full); err != nil {
		return nil, fmt.Errorf("error decoding op full response: %w", err)
	}
	return &full, nil
}

// createItem creates an empty item in the vault based on the project name
// op item create --category Password --vault FTL --title ftl.projectname.secrets
func (o OnePasswordProvider) createItem(ctx context.Context, vault string) error {
	args := []string{
		"item", "create",
		"--category", "Password",
		"--vault", vault,
		"--title", o.itemName(),
	}
	_, err := exec.Capture(ctx, ".", "op", args...)
	if err != nil {
		return fmt.Errorf("create item failed in vault %q: %w", vault, err)
	}
	return nil
}

// op item edit 'ftl.projectname.secrets' 'module.secretname[password]=value with space'
func (o OnePasswordProvider) storeSecret(ctx context.Context, vault string, ref Ref, secret []byte) error {
	module, ok := ref.Module.Get()
	if !ok {
		return fmt.Errorf("module is required for secret: %v", ref)
	}
	args := []string{
		"item", "edit", o.itemName(),
		"--vault", vault,
		fmt.Sprintf("username[text]=%s", defaultSecretModificationWarning),
		fmt.Sprintf("%s\\.%s[password]=%s", module, ref.Name, string(secret)),
	}
	_, err := exec.Capture(ctx, ".", "op", args...)
	if err != nil {
		return fmt.Errorf("edit item failed in vault %q, ref %q: %w", vault, ref, err)
	}
	return nil
}
