package model

import (
	"math/rand"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
)

// Ensure key randomness is deterministic
func ensureDeterministicRand(t *testing.T) {
	t.Helper()
	oldRandRead := randRead
	t.Cleanup(func() { randRead = oldRandRead })
	randRead = rand.New(rand.NewSource(0)).Read //nolint:gosec
}

func TestRunnerKey(t *testing.T) {
	ensureDeterministicRand(t)
	for _, test := range []struct {
		key         RunnerKey
		expected    RunnerKey
		expectedStr string
	}{
		// Production Keys
		{key: NewRunnerKey("0.0.0.0", "8080"),
			expected: RunnerKey{
				Payload: RunnerPayload{HostPortMixin{Hostname: optional.Some("0.0.0.0"), Port: "8080"}},
				Suffix:  []byte("\x01\x94\xfd\xc2\xfa/\xfc\xc0A\xd3"),
			},
			expectedStr: "rnr-0.0.0.0-8080-17snepfuemu5iab"},
		{key: NewRunnerKey("example-host-with-hyphens", "0"),
			expectedStr: "rnr-example-host-with-hyphens-0-5g5cadeqxpqe574v"},
		{key: NewRunnerKey("noport", ""),
			expectedStr: "rnr-noport--59gwlv6lkyexwxf1"},
		{key: NewRunnerKey("rnr-hostwithsameprefix", "80"),
			expectedStr: "rnr-rnr-hostwithsameprefix-80-8ta4kn0wr7k2mlc"},
		{key: NewRunnerKey("rnr-hostwithprefixandfakeport-80", "80"),
			expected: RunnerKey{
				Payload: RunnerPayload{HostPortMixin{Hostname: optional.Some("rnr-hostwithprefixandfakeport-80"), Port: "80"}},
				Suffix:  []byte("\xb1\r9FQ\x85\x0fԡx"),
			},
			expectedStr: "rnr-rnr-hostwithprefixandfakeport-80-80-3s5h946y6kylrxx4"},

		// // Local Keys
		{key: NewLocalRunnerKey(0), expectedStr: "rnr-0-2xhrhnc74bsx9rp4"},
		{key: NewLocalRunnerKey(1), expectedStr: "rnr-1-6i7pgd8mri5ti7f"},
	} {
		key := test.key
		t.Run(test.key.String(), func(t *testing.T) {
			parsed, err := ParseRunnerKey(key.String())
			assert.NoError(t, err)
			assert.Equal(t, key, parsed, "expected %v for %v after parsing", key, parsed)

			assert.Equal(t, test.expectedStr, key.String())

			anyValue, err := key.Value()
			assert.NoError(t, err)
			value, ok := anyValue.(string)
			assert.True(t, ok, "expected string value for key %v", key)
			assert.Equal(t, test.expectedStr, value)

			var scanKey RunnerKey
			err = scanKey.Scan(value)
			assert.NoError(t, err)
			assert.Equal(t, key, scanKey)

			var zero RunnerKey
			if !assert.Compare(t, zero, test.expected) {
				assert.Equal(t, test.expected, key)
			}
		})
	}
}
