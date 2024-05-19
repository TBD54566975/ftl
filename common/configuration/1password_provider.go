package configuration

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/TBD54566975/ftl/internal/slices"
	"net/url"
	"regexp"
	"strings"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/exec"
	"github.com/TBD54566975/ftl/internal/log"
)

// OnePasswordProvider is a configuration provider that reads passwords from
// 1Password vaults via the "op" command line tool.
type OnePasswordProvider struct {
	OnePassword bool `name:"op" help:"Write 1Password secret references - does not write to 1Password." group:"Provider:" xor:"configwriter"`
}

var _ MutableProvider[Secrets] = OnePasswordProvider{}

func (OnePasswordProvider) Role() Secrets                               { return Secrets{} }
func (o OnePasswordProvider) Key() string                               { return "op" }
func (o OnePasswordProvider) Delete(ctx context.Context, ref Ref) error { return nil }

func (o OnePasswordProvider) Load(ctx context.Context, ref Ref, key *url.URL) ([]byte, error) {
	logger := log.FromContext(ctx)

	_, err := exec.LookPath("op")
	if err != nil {
		return nil, fmt.Errorf("1Password CLI tool \"op\" not found: %w", err)
	}

	decoded, err := base64.RawURLEncoding.DecodeString(key.Host)
	if err != nil {
		return nil, fmt.Errorf("1Password secret reference must be a base64 encoded string: %w", err)
	}

	parsedRef, err := decodeSecretRef(string(decoded))
	if err != nil {
		return nil, fmt.Errorf("1Password secret reference invalid: %w", err)
	}

	//output, err := exec.Capture(ctx, ".", "op", "read", "-n", string(decoded))
	//if err != nil {
	//	lines := bytes.Split(output, []byte("\n"))
	//	logger := log.FromContext(ctx)
	//	for _, line := range lines {
	//		logger.Warnf("%s", line)
	//	}
	//	return nil, fmt.Errorf("error running 1password CLI tool \"op\": %w", err)
	//}

	// A single password: op --format json item get --vault Personal "With Spaces" --fields=username
	// { id, value }
	// All fields:        op --format json item get --vault Personal "With Spaces"
	// { fields: [ { id, value } ] }

	args := []string{"--format", "json", "item", "get", "--vault", parsedRef.Vault, parsedRef.Item}

	v, fieldSpecified := parsedRef.Field.Get()
	if fieldSpecified {
		args = append(args, "--fields", v)
	}

	output, err := exec.Capture(ctx, ".", "op", args...)
	if err != nil {
		return nil, fmt.Errorf("run `op` with args %v: %w", args, err)
	}

	logger.Debugf("output: %s", output)
	logger.Debugf("fieldSpecified: %v", fieldSpecified)

	if fieldSpecified {
		v, err := decodeSingleResponse(output)
		if err != nil {
			return nil, err
		}
		logger.Debugf("decoed v: %v", v)

		return json.Marshal(v.Value)
	} else {
		v, err := decodeFullResponse(output)
		if err != nil {
			return nil, err
		}

		// Filter out anything without a value
		filtered := slices.Filter(v, func(e entry) bool {
			return e.Value != ""
		})
		// Map to id: value
		var mapped = make(map[string]string)
		for _, e := range filtered {
			mapped[e.Id] = e.Value
		}

		return json.Marshal(mapped)
	}
}

func (o OnePasswordProvider) Store(ctx context.Context, ref Ref, value []byte) (*url.URL, error) {
	var opref string
	if err := json.Unmarshal(value, &opref); err != nil {
		return nil, fmt.Errorf("1Password value must be a JSON string containing a 1Password secret refererence: %w", err)
	}
	if !strings.HasPrefix(opref, "op://") {
		return nil, fmt.Errorf("1Password secret reference must start with \"op://\"")
	}

	encoded := base64.RawURLEncoding.EncodeToString([]byte(opref))
	return &url.URL{Scheme: "op", Host: encoded}, nil
}

func (o OnePasswordProvider) Writer() bool { return o.OnePassword }

type entry struct {
	Id    string `json:"id"`
	Value string `json:"value"`
}

type full struct {
	Fields []entry `json:"fields"`
}

// Decode a full response from op
func decodeFullResponse(output []byte) ([]entry, error) {
	var full full
	if err := json.Unmarshal(output, &full); err != nil {
		return nil, fmt.Errorf("error decoding op full response: %w", err)
	}
	return full.Fields, nil
}

// Decode a single response from op
func decodeSingleResponse(output []byte) (*entry, error) {
	var single entry
	if err := json.Unmarshal(output, &single); err != nil {
		return nil, fmt.Errorf("error decoding op single response: %w", err)
	}
	return &single, nil
}

// Custom parser for 1Password secret references because the format is not a standard URL, and we also need to
// allow users to omit the field name so that we can support secrets with multiple fields.
//
// Does not support "section-name".
//
//	op://<vault-name>/<item-name>[/<field-name>]
//
// Secret references are case-insensitive and support the following characters:
//
//	alphanumeric characters (a-z, A-Z, 0-9), -, _, . and the whitespace character
//
// If an item or field name includes a / or an unsupported character, use the item
// or field's unique identifier (ID) instead of its name.
//
// See https://developer.1password.com/docs/cli/secrets-reference-syntax/
type secretRef struct {
	Vault string
	Item  string
	Field optional.Option[string]
}

var validCharsRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_\. ]+$`)

func decodeSecretRef(ref string) (*secretRef, error) {

	// Take out and check the "op://" prefix
	const prefix = "op://"
	if !strings.HasPrefix(ref, prefix) {
		return nil, fmt.Errorf("must start with \"op://\"")
	}
	ref = ref[len(prefix):]

	parts := strings.Split(ref, "/")

	if len(parts) < 2 {
		return nil, fmt.Errorf("must have at least 2 parts")
	}
	if len(parts) > 3 {
		return nil, fmt.Errorf("must have at most 3 parts")
	}

	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("url parts must not be empty")
		}

		if !validCharsRegex.MatchString(part) {
			return nil, fmt.Errorf("url part %q contains unsupported characters. regex: %q", part, validCharsRegex)
		}
	}

	secret := secretRef{
		Vault: parts[0],
		Item:  parts[1],
		Field: optional.None[string](),
	}
	if len(parts) == 3 {
		secret.Field = optional.Some(parts[2])
	}

	return &secret, nil
}
