package main

import (
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestLongestCommonPrefix(t *testing.T) {
	foo := []string{"foo/waz/boo", "foo/waz/bar", "foo/waz/baz"}
	assert.Equal(t, "foo/waz", longestCommonPathPrefix(foo))
}
