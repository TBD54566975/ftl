package bind

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/alecthomas/atomic"
)

type BindAllocator struct {
	baseURL           *url.URL
	dynamicPortsAfter int
	port              atomic.Int32
}

// NewBindAllocator creates a BindAllocator, which dynamically allocates ports for binding local servers.
//
// "staticPorts" is the number of ports that are statically allocated and that must be free before dynamic ports
// are allocated.
func NewBindAllocator(url *url.URL, staticPorts int) (*BindAllocator, error) {
	_, portStr, err := net.SplitHostPort(url.Host)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	return &BindAllocator{
		baseURL:           url,
		port:              atomic.NewInt32(int32(port) - 1), //nolint:gosec
		dynamicPortsAfter: port + staticPorts,
	}, nil
}

func (b *BindAllocator) NextPort() (int, error) {
	var l *net.TCPListener
	var err error

	maxTries := 5000

	tries := 0
	for {
		tries++
		port := int(b.port.Add(1))
		l, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP(b.baseURL.Hostname()), Port: port})

		if err != nil {
			if port < b.dynamicPortsAfter {
				return 0, fmt.Errorf("failed to bind to port %d: %w", port, err)
			}
			if tries >= maxTries {
				panic("failed to find an open port: " + err.Error())
			}
			continue
		}
		_ = l.Close()

		return port, nil
	}
}

func (b *BindAllocator) Next() (*url.URL, error) {
	newURL := *b.baseURL
	nextPort, err := b.NextPort()
	if err != nil {
		return nil, err
	}
	newURL.Host = net.JoinHostPort(b.baseURL.Hostname(), strconv.Itoa(nextPort))
	return &newURL, nil
}
