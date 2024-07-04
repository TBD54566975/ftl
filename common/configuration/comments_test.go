package configuration

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestComments(t *testing.T) {
	for _, tt := range []struct {
		input   string
		comment string
		output  string
	}{
		{
			input:   "test input can be anything",
			comment: "This is a test",
			output:  "# This is a test\ntest input can be anything",
		},
		{
			input:   "{\n  \"key\": \"value\"\n}",
			comment: "This is a multi\nline\ncomment",
			output:  "# This is a multi\n# line\n# comment\n{\n  \"key\": \"value\"\n}",
		},
	} {
		t.Run(tt.input, func(t *testing.T) {
			wrapped := wrapWithComments([]byte(tt.input), tt.comment)
			assert.Equal(t, tt.output, string(wrapped))

			unwrapped := unwrapComments(wrapped)
			assert.Equal(t, tt.input, string(unwrapped))
		})
	}
}
