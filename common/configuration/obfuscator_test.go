package configuration

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/alecthomas/types/optional"
)

func TestObfuscator(t *testing.T) {
	defaultKey := []byte("1234567890123456") // 32 characters
	for _, tt := range []struct {
		input               string
		key                 []byte
		expectedError       optional.Option[string]
		backwardsCompatible bool
	}{
		{
			input:               "test input can be anything",
			key:                 defaultKey,
			backwardsCompatible: false,
		},
		{
			input:               `"test input can be anything"`,
			key:                 defaultKey,
			backwardsCompatible: true,
		},
		{
			input:               `"{\n  "key": "value"\n}`,
			key:                 defaultKey,
			backwardsCompatible: true,
		},
		{
			input:               `1.2345`,
			key:                 defaultKey,
			backwardsCompatible: false,
		},
		{
			input:         "key is too short",
			key:           []byte("too-short"),
			expectedError: optional.Some("could not create cypher for obfuscation: crypto/aes: invalid key size 9"),
		},
	} {
		t.Run(tt.input, func(t *testing.T) {
			o := Obfuscator{
				key: tt.key,
			}
			// obfuscate
			obfuscated, err := o.Obfuscate([]byte(tt.input))
			if expectedError, ok := tt.expectedError.Get(); ok {
				assert.EqualError(t, err, expectedError)
				return
			}
			assert.NoError(t, err)

			// reveal obfuscated value
			revealed, err := o.Reveal(obfuscated)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, string(revealed))

			// obfuscated value should not include the input we are trying to obfuscate
			assert.NotContains(t, string(obfuscated), tt.input)

			// reveal unobfuscated value to check backwards compatibility
			if tt.backwardsCompatible {
				revealed, err = o.Reveal([]byte(tt.input))
				assert.NoError(t, err)
				assert.Equal(t, tt.input, string(revealed))
			}
		})
	}
}
