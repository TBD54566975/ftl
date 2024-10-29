package main

import (
	"net/url"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestBindLocalWithRemoteEndpoint(t *testing.T) {
	var err error
	cli.Endpoint, err = url.Parse("http://block.xyz:80")
	assert.NoError(t, err)

	bindAllocator, err := bindAllocatorWithoutController()
	assert.NoError(t, err)

	nextUrl, err := bindAllocator.Next()
	assert.NoError(t, err)

	assert.True(t, strings.HasPrefix(nextUrl.String(), "http://127.0.0.1:"), nextUrl.String())

	bindAllocatorWithoutController()
}
