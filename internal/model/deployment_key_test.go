package model

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestDeploymentKey(t *testing.T) {
	for _, test := range []struct {
		str       string // when full string is known
		strPrefix string // when only prefix is known
		module    string
		hash      string
		decodeErr bool
	}{
		{module: "time", strPrefix: "time-"},
		{str: "time-00112233", decodeErr: true},
		{str: "time-001122334455", decodeErr: true},
		{str: "time-0011223344", module: "time", hash: "0011223344"},
		{str: "-0011223344", decodeErr: true},
		{str: "module-with-hyphens-0011223344", module: "module-with-hyphens", hash: "0011223344"},
		{str: "-", decodeErr: true},
	} {
		decoded, decodeErr := ParseDeploymentKey(test.str)

		if test.decodeErr {
			assert.Error(t, decodeErr, "expected error for deployment key %q", test.str)
		} else {
			created := NewDeploymentKey(test.module)

			forceEncoded := DeploymentKey{
				module: test.module,
				hash:   test.hash,
			}

			if test.str != "" && test.module != "" {
				assert.Equal(t, test.module, decoded.module, "expected module %q for %q", test.module, decoded.module)
			}

			if test.str != "" && test.hash != "" {
				assert.Equal(t, test.hash, decoded.hash, "expected hash %q for %q", test.hash, decoded.hash)
			}

			if test.module != "" && test.strPrefix != "" {
				assert.True(t, strings.HasPrefix(created.String(), test.strPrefix), "expected string prefix %q for %q", test.strPrefix, created.String())
			}

			if test.module != "" && test.hash != "" && test.str != "" {
				assert.Equal(t, test.str, forceEncoded.String(), "expected string %q for %q", test.str, forceEncoded.String())
			}
		}
	}
}

func TestZeroDeploymentKey(t *testing.T) {
	_, err := ParseDeploymentKey("")
	assert.Error(t, err, "expected error for empty deployment key")
}
