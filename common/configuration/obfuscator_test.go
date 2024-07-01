package configuration

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
)

func TestObfuscator(t *testing.T) {
	defaultKey := []byte("1234567890123456") // 32 characters
	for _, tt := range []struct {
		input            string
		comment          string
		obfuscatedPrefix string
		key              []byte
		expectedError    optional.Option[string]
	}{
		{
			input:            "test input can be anything",
			comment:          "This is a test",
			obfuscatedPrefix: "# This is a test\n",
			key:              defaultKey,
		},
		{
			input:            "{\n  \"key\": \"value\"\n}",
			comment:          "This is a multi\nline\ncomment",
			obfuscatedPrefix: "# This is a multi\n# line\n# comment\n",
			key:              defaultKey,
		},
		{
			input:         "key is too short",
			key:           []byte("too-short"),
			expectedError: optional.Some("crypto/aes: invalid key size 9"),
		},
	} {
		t.Run(tt.input, func(t *testing.T) {
			o := Obfuscator{
				key: tt.key,
			}
			obfuscated, err := o.Obfuscate([]byte(tt.input), tt.comment)
			if expectedError, ok := tt.expectedError.Get(); ok {
				assert.EqualError(t, err, expectedError)
				return
			}
			assert.NoError(t, err)
			assert.HasPrefix(t, string(obfuscated), tt.obfuscatedPrefix)
			revealed, err := o.Reveal(obfuscated)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, string(revealed))

			// obfuscated value should not include the input we are trying to obfuscate
			assert.NotContains(t, string(obfuscated), tt.input)
		})
	}
}
