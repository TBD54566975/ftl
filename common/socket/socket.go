// Package socket is a type to represent a (network, addr) pair.
package socket

import (
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/alecthomas/errors"
)

// Socket represents a (network, addr) pair. Its serialised form is a URI:
// tcp://127.0.0.1:8080 or unix:///tmp/foo.sock
type Socket struct {
	Network string
	Addr    string
}

func (s Socket) URL() *url.URL {
	return &url.URL{Scheme: "http", Host: s.Addr}
}

// Valid returns true if the Socket is valid.
func (s Socket) Valid() bool {
	return s.Network != "" && s.Addr != ""
}

func (s Socket) String() string {
	return fmt.Sprintf("%s://%s", s.Network, s.Addr)
}

func (s *Socket) UnmarshalText(b []byte) error {
	tmp, err := Parse(string(b))
	if err != nil {
		return errors.WithStack(err)
	}
	*s = tmp
	return nil
}

// Listen on a socket.
//
// For unix sockets, the socket will be removed if it already exists.
func Listen(s Socket) (net.Listener, error) {
	if s.Network == "unix" {
		if err := os.Remove(s.Addr); err != nil && !os.IsNotExist(err) {
			return nil, errors.WithStack(err)
		}
	}
	l, err := net.Listen(s.Network, s.Addr)
	return l, errors.WithStack(err)
}

// Parse a socket URI into a network and address.
//
// Supported URI schemes are "unix://<path>" and "tcp://<host>:<port>"
func Parse(uri string) (Socket, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return Socket{}, errors.WithStack(err)
	}
	switch u.Scheme {
	case "unix":
		return Socket{Network: "unix", Addr: u.Path}, nil
	case "tcp":
		return Socket{Network: "tcp", Addr: u.Host}, nil
	}
	return Socket{}, errors.Errorf("unsupported socket URI %q", uri)
}

// MustParse is like Parse but panics on error.
func MustParse(uri string) Socket {
	s, err := Parse(uri)
	if err != nil {
		panic(err)
	}
	return s
}
