package configuration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

// OnePasswordProvider is a configuration provider that reads passwords from
// 1Password vaults via the "op" command line tool.
type OnePasswordProvider struct {
	OnePassword bool `name:"op" help:"Write 1Password secret references - does not write to 1Password." group:"Provider:" xor:"configwriter"`
}

var _ MutableProvider = OnePasswordProvider{}

func (o OnePasswordProvider) Key() string                               { return "op" }
func (o OnePasswordProvider) Delete(ctx context.Context, ref Ref) error { return nil }

func (o OnePasswordProvider) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	_, err := exec.LookPath("op")
	if err != nil {
		return nil, fmt.Errorf("1Password CLI tool \"op\" not found: %w", err)
	}
	output, err := exec.Capture(ctx, ".", "op", "read", "-n", key.String())
	if err != nil {
		lines := bytes.Split(output, []byte("\n"))
		logger := log.FromContext(ctx)
		for _, line := range lines {
			logger.Warnf("%s", line)
		}
		return nil, fmt.Errorf("error running 1password CLI tool \"op\": %w", err)
	}
	return json.Marshal(string(output))
}

func (o OnePasswordProvider) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	var opref string
	if err := json.Unmarshal(value, &opref); err != nil {
		return nil, fmt.Errorf("1Password value must be a JSON string containing a 1Password secret refererence: %w", err)
	}
	u, err := url.Parse(opref)
	if err != nil {
		return nil, fmt.Errorf("invalid 1Password item ID: %w", err)
	}
	return u, nil
}

func (o OnePasswordProvider) Writer() bool { return o.OnePassword }
