package configuration

import (
	"strings"

	"github.com/TBD54566975/ftl/internal/slices"
)

const defaultSecretModificationWarning = `This secret is managed by "ftl secret set", DO NOT MODIFY`

// wrapWithComments wraps the secret with a comment to indicate that it is managed by FTL.
//
// This is used by providers that want to include a warning to avoid manual modification.
// The provider must support multiline secrets.
// Comment lines are prefixed with '# ' in the result.
func wrapWithComments(secret []byte, comments string) []byte {
	lines := []string{}
	for _, line := range strings.Split(comments, "\n") {
		lines = append(lines, "# "+line)
	}
	lines = append(lines, string(secret))
	return []byte(strings.Join(lines, "\n"))
}

// unwrapComments removes comments if they exist by looking for the lines starting with '#'
func unwrapComments(secret []byte) []byte {
	lines := strings.Split(string(secret), "\n")
	lines = slices.Filter(lines, func(line string) bool {
		return !strings.HasPrefix(line, "#")
	})
	return []byte(strings.Join(lines, "\n"))
}
