package main

import (
	"fmt"
	"net"
	"net/url"

	"github.com/TBD54566975/ftl/internal/bind"
)

// Create a bind allocator that skips the reserved port for the controller.
//
// The bind allocator will use cli.Endpoint if it is a local URL with a port. Otherwise a default port is used.
func bindAllocatorWithoutController() (*bind.BindAllocator, error) {
	var url *url.URL
	var err error
	// use cli.Endpoint if it is a local URL with a port
	if cli.Endpoint != nil && cli.Endpoint.Port() != "" {
		h := cli.Endpoint.Hostname()
		ips, err := net.LookupIP(h)
		if err != nil {
			return nil, fmt.Errorf("failed to look up IP: %w", err)
		}
		for _, netip := range ips {
			if netip.IsLoopback() {
				url = cli.Endpoint
				break
			}
		}
	}
	// fallback to default
	if url == nil {
		url, err = url.Parse("http://127.0.0.1:8892")
		if err != nil {
			return nil, fmt.Errorf("failed to parse default URL: %w", err)
		}
	}
	bindAllocator, err := bind.NewBindAllocator(url, 0)
	if err != nil {
		return nil, fmt.Errorf("could not create bind allocator: %w", err)
	}
	// Skip initial port as it is reserved for the controller
	_, _ = bindAllocator.Next() //nolint:errcheck
	return bindAllocator, nil
}
