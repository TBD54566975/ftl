package bind

import (
	"net"
	"net/url"
	"strconv"

	"github.com/alecthomas/atomic"
)

type BindAllocator struct {
	baseURL *url.URL
	port    atomic.Int32
}

func NewBindAllocator(url *url.URL) (*BindAllocator, error) {
	_, portStr, err := net.SplitHostPort(url.Host)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	return &BindAllocator{
		baseURL: url,
		port:    atomic.NewInt32(int32(port) - 1), //nolint:gosec
	}, nil
}

func (b *BindAllocator) NextPort() int {
	var l *net.TCPListener
	var err error

	maxTries := 5000

	tries := 0
	for {
		tries++
		port := int(b.port.Add(1))
		l, err = net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP(b.baseURL.Hostname()), Port: port})

		if err != nil {
			if tries >= maxTries {
				panic("failed to find an open port: " + err.Error())
			}
			continue
		}
		_ = l.Close()

		return port
	}
}

func (b *BindAllocator) Next() *url.URL {
	newURL := *b.baseURL
	newURL.Host = net.JoinHostPort(b.baseURL.Hostname(), strconv.Itoa(b.NextPort()))
	return &newURL
}
