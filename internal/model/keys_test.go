package model

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestRunnerKey(t *testing.T) {
	for _, test := range []struct {
		key         RunnerKey
		str         string
		strPrefix   string
		value       string
		valuePrefix string
	}{
		// Production Keys
		{key: NewRunnerKey("0.0.0.0", "8080"), strPrefix: "r-0.0.0.0-8080-", valuePrefix: "0.0.0.0-8080-"},
		{key: NewRunnerKey("example-host-with-hyphens", "0"), strPrefix: "r-example-host-with-hyphens-0-", valuePrefix: "example-host-with-hyphens-0-"},
		{key: NewRunnerKey("noport", ""), strPrefix: "r-noport--", valuePrefix: "noport--"},
		{key: NewRunnerKey("r-hostwithsameprefix", "80"), strPrefix: "r-r-hostwithsameprefix-80-", valuePrefix: "r-hostwithsameprefix-80-"},
		{key: NewRunnerKey("r-hostwithprefixandfakeport-80", "80"), strPrefix: "r-r-hostwithprefixandfakeport-80-80-", valuePrefix: "r-hostwithprefixandfakeport-80-80-"},

		// Local Keys
		{key: NewLocalRunnerKey(0), str: "r-0000", value: "0000"},
		{key: NewLocalRunnerKey(1), str: "r-0001", value: "0001"},
		{key: NewLocalRunnerKey(9999), str: "r-9999", value: "9999"},
		{key: NewLocalRunnerKey(12345), str: "r-12345", value: "12345"},
	} {
		if test.str != "" {
			assert.Equal(t, test.str, test.key.String(), "expected string %q for %q", test.str, test.key.String())
		}
		if test.strPrefix != "" {
			assert.True(t, strings.HasPrefix(test.key.String(), test.strPrefix), "expected string prefix %q for %q", test.strPrefix, test.key.String())
		}
		aValue, err := test.key.Value()
		assert.NoError(t, err)
		value := aValue.(string)

		if test.value != "" {
			assert.Equal(t, test.value, value, "expected value %q for %q", test.value, value)
		}
		if test.valuePrefix != "" {
			assert.True(t, strings.HasPrefix(value, test.valuePrefix), "expected value prefix %q for %q", test.valuePrefix, value)
		}

		parsed, err := ParseRunnerKey(test.key.String())
		assert.NoError(t, err)
		assert.Equal(t, test.key, parsed, "expected %v for %v after parsing", test.key, parsed)

		parsed, err = parseKey[RunnerKey](value, false)
		assert.NoError(t, err)
		assert.Equal(t, test.key, parsed, "expected %v for %v after parsing db key", test.key, parsed)
	}
}
