package configuration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/atomic"
	"github.com/alecthomas/types/optional"

	"github.com/kballard/go-shellquote"

	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
	"github.com/TBD54566975/ftl/internal/slices"
)

const (
	OnePasswordKey = "op"
)

// OnePasswordProvider is a configuration provider that reads passwords from
// 1Password vaults via the "op" command line tool.
type OnePasswordProvider struct {
	Vault string

	// When 1Password is locked we don't want to bring up multiple prompts.
	// The following approach tries to balance throughput while minimizing multiple prompts:
	// - If a successful call to 1Password was completed within the past 1 second, we allow concurrent calls to 1Password
	// - Otherwise we make a single call to 1Password, coordinated via the lock.
	lock          sync.Mutex
	latestSuccess atomic.Value[optional.Option[time.Time]]
}

func NewOnePasswordProvider(vault string) *OnePasswordProvider {
	return &OnePasswordProvider{
		Vault: vault,
		// concurrentModules: xsync.NewMapOf[string, bool](),
	}
}

func (*OnePasswordProvider) Role() Secrets { return Secrets{} }
func (o *OnePasswordProvider) Key() string { return OnePasswordKey }
func (o *OnePasswordProvider) Delete(ctx context.Context, ref Ref) error {
	return nil
}

func (o *OnePasswordProvider) canAccessWithoutLock() bool {
	latestSuccess, ok := o.latestSuccess.Load().Get()
	return ok && time.Since(latestSuccess) <= time.Second
}

func safeAccess[T any](o *OnePasswordProvider, f func() (T, error)) (result T, err error) {
	if err := checkOpBinary(); err != nil {
		return result, err
	}
	if !o.canAccessWithoutLock() {
		o.lock.Lock()
		if o.canAccessWithoutLock() {
			// immediately unlock so other access can also proceed
			o.lock.Unlock()
		} else {
			// exclusive access to 1Password
			defer o.lock.Unlock()
		}
	}
	result, err = f()
	if err == nil {
		o.latestSuccess.Store(optional.Some[time.Time](time.Now()))
	}
	return
}

// Load returns the secret stored in 1password.
func (o *OnePasswordProvider) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	return safeAccess(o, func() ([]byte, error) {
		vault := key.Host
		full, err := getItem(ctx, vault, ref)
		if err != nil {
			return nil, fmt.Errorf("get item failed: %w", err)
		}

		secret, ok := full.password()
		if !ok {
			return nil, fmt.Errorf("password field not found in item %q", ref)
		}

		// Just to verify that it is JSON encoded.
		var decoded interface{}
		err = json.Unmarshal(secret, &decoded)
		if err != nil {
			return nil, fmt.Errorf("secret is not JSON encoded: %w", err)
		}

		return secret, nil
	})
}

var vaultRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-.]+$`)

// Store will save the given secret in 1Password via the `op` command.
//
// op does not support "create or update" as a single command. Neither does it support specifying an ID on create.
// Because of this, we need check if the item exists before creating it, and update it if it does.
func (o *OnePasswordProvider) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	return safeAccess(o, func() (*url.URL, error) {
		if o.Vault == "" {
			return nil, fmt.Errorf("vault missing, specify vault as a flag to the controller")
		}
		if !vaultRegex.MatchString(o.Vault) {
			return nil, fmt.Errorf("vault name %q contains invalid characters. a-z A-Z 0-9 _ . - are valid", o.Vault)
		}

		url := &url.URL{Scheme: "op", Host: o.Vault}

		_, err := getItem(ctx, o.Vault, ref)
		if errors.As(err, new(itemNotFoundError)) {
			err = createItem(ctx, o.Vault, ref, value)
			if err != nil {
				return nil, fmt.Errorf("create item failed: %w", err)
			}
			return url, nil

		} else if err != nil {
			return nil, fmt.Errorf("get item failed: %w", err)
		}

		err = editItem(ctx, o.Vault, ref, value)
		if err != nil {
			return nil, fmt.Errorf("edit item failed: %w", err)
		}

		return url, nil
	})
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
	ref   Ref
}

func (e itemNotFoundError) Error() string {
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

func (i item) password() ([]byte, bool) {
	secret, ok := slices.Find(i.Fields, func(item entry) bool {
		return item.ID == "password"
	})
	return []byte(secret.Value), ok
}

// op --format json item get --vault Personal "With Spaces"
func getItem(ctx context.Context, vault string, ref Ref) (*item, error) {
	logger := log.FromContext(ctx)

	args := []string{"--format", "json", "item", "get", "--vault", vault, ref.String()}
	output, err := exec.Capture(ctx, ".", "op", args...)
	logger.Debugf("Getting item with args %s", shellquote.Join(args...))
	if err != nil {
		// This is specifically not itemNotFoundError, to distinguish between vault not found and item not found.
		if strings.Contains(string(output), "isn't a vault") {
			return nil, fmt.Errorf("vault %q not found: %w", vault, err)
		}

		// Item not found, seen two ways of reporting this:
		if strings.Contains(string(output), "not found in vault") {
			return nil, itemNotFoundError{vault, ref}
		}
		if strings.Contains(string(output), "isn't an item") {
			return nil, itemNotFoundError{vault, ref}
		}

		return nil, fmt.Errorf("run `op` with args %s: %w", shellquote.Join(args...), err)
	}

	var full item
	if err := json.Unmarshal(output, &full); err != nil {
		return nil, fmt.Errorf("error decoding op full response: %w", err)
	}
	return &full, nil
}

// op item create --category Password --vault FTL --title mod.ule "password=val ue"
func createItem(ctx context.Context, vault string, ref Ref, secret []byte) error {
	args := []string{"item", "create", "--category", "Password", "--vault", vault, "--title", ref.String(), "password=" + string(secret)}
	_, err := exec.Capture(ctx, ".", "op", args...)
	if err != nil {
		return fmt.Errorf("create item failed in vault %q, ref %q: %w", vault, ref, err)
	}

	return nil
}

// op item edit --vault ftl test "password=with space"
func editItem(ctx context.Context, vault string, ref Ref, secret []byte) error {
	args := []string{"item", "edit", "--vault", vault, ref.String(), "password=" + string(secret)}
	_, err := exec.Capture(ctx, ".", "op", args...)
	if err != nil {
		return fmt.Errorf("edit item failed in vault %q, ref %q: %w", vault, ref, err)
	}

	return nil
}
