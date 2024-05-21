package configuration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/slices"
)

// OnePasswordProvider is a configuration provider that reads passwords from
// 1Password vaults via the "op" command line tool.
type OnePasswordProvider struct {
	Vault string `name:"op" help:"Store a secret in this 1Password vault." group:"Provider:" xor:"configwriter" placeholder:"VAULT"`
}

var _ MutableProvider[Secrets] = OnePasswordProvider{}

func (OnePasswordProvider) Role() Secrets                               { return Secrets{} }
func (o OnePasswordProvider) Key() string                               { return "op" }
func (o OnePasswordProvider) Delete(ctx context.Context, ref Ref) error { return nil }

// Load returns the secret stored in 1password, quoted as a JSON string.
func (o OnePasswordProvider) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	if err := checkOpBinary(); err != nil {
		return nil, err
	}

	vault := key.Host
	full, err := getItem(ctx, vault, ref)
	if err != nil {
		return nil, fmt.Errorf("get item failed: %w", err)
	}

	secret, ok := slices.Find(full.Fields, func(item entry) bool {
		return item.ID == "password"
	})
	if !ok {
		return nil, fmt.Errorf("password field not found in item %q", ref)
	}

	jsonSecret, err := json.Marshal(secret.Value)
	if err != nil {
		return nil, fmt.Errorf("json marshal failed: %w", err)
	}

	return jsonSecret, nil
}

// Store will save the given secret in 1Password via the `op` command.
//
// op does not support "create or update" as a single command. Neither does it support specifying an ID on create.
// Because of this, we need check if the item exists before creating it, and update it if it does.
func (o OnePasswordProvider) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	if err := checkOpBinary(); err != nil {
		return nil, err
	}

	var secret string
	err := json.Unmarshal(value, &secret)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w", err)
	}

	url := &url.URL{Scheme: "op", Host: o.Vault}

	_, err = getItem(ctx, o.Vault, ref)
	var notFound notFoundError
	if errors.As(err, &notFound) {
		err = createItem(ctx, o.Vault, ref, secret)
		if err != nil {
			return nil, fmt.Errorf("create item failed: %w", err)
		}
		return url, nil

	} else if err != nil {
		return nil, fmt.Errorf("get item failed: %w", err)
	}

	err = editItem(ctx, o.Vault, ref, secret)
	if err != nil {
		return nil, fmt.Errorf("edit item failed: %w", err)
	}

	return url, nil
}

func (o OnePasswordProvider) Writer() bool { return o.Vault != "" }

func checkOpBinary() error {
	_, err := exec.LookPath("op")
	if err != nil {
		return fmt.Errorf("1Password CLI tool \"op\" not found: %w", err)
	}
	return nil
}

type notFoundError struct {
	vault string
	ref   Ref
}

func (e notFoundError) Error() string {
	return fmt.Sprintf("item %q not found in vault %q", e.ref, e.vault)
}

// item is the JSON response from `op item get`.
type item struct {
	Fields []entry `json:"fields"`
}

type entry struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

// op --format json item get --vault Personal "With Spaces"
func getItem(ctx context.Context, vault string, ref Ref) (*item, error) {
	logger := log.FromContext(ctx)

	args := []string{"--format", "json", "item", "get", "--vault", vault, ref.String()}
	output, err := exec.Capture(ctx, ".", "op", args...)
	logger.Debugf("Getting item with args %v", args)
	if err != nil {
		if strings.Contains(string(output), "isn't a vault") {
			return nil, fmt.Errorf("vault %q not found: %w", vault, err)
		}

		// Item not found, seen two ways of reporting this:
		if strings.Contains(string(output), "not found in vault") {
			return nil, notFoundError{vault, ref}
		}
		if strings.Contains(string(output), "isn't an item") {
			return nil, notFoundError{vault, ref}
		}

		return nil, fmt.Errorf("run `op` with args %v: %w", args, err)
	}

	var full item
	if err := json.Unmarshal(output, &full); err != nil {
		return nil, fmt.Errorf("error decoding op full response: %w", err)
	}
	return &full, nil
}

// op item create --category Password --vault FTL --title mod.ule "password=val ue"
func createItem(ctx context.Context, vault string, ref Ref, secret string) error {
	args := []string{"item", "create", "--category", "Password", "--vault", vault, "--title", ref.String(), "password=" + secret}
	_, err := exec.Capture(ctx, ".", "op", args...)
	if err != nil {
		return fmt.Errorf("create item failed in vault %q, ref %q: %w", vault, ref, err)
	}

	return nil
}

// op item edit --vault ftl test "password=with space"
func editItem(ctx context.Context, vault string, ref Ref, secret string) error {
	args := []string{"item", "edit", "--vault", vault, ref.String(), "password=" + secret}
	_, err := exec.Capture(ctx, ".", "op", args...)
	if err != nil {
		return fmt.Errorf("edit item failed in vault %q, ref %q: %w", vault, ref, err)
	}

	return nil
}
