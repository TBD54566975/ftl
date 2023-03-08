package socket

import (
	"context"
	"net"
	"net/url"
	"os"

	"github.com/alecthomas/errors"
)

type Socket struct {
	Network string
	Addr    string
}

func (s Socket) Dial(ctx context.Context) (net.Conn, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, s.Network, s.Addr)
	return conn, errors.WithStack(err)
}

// Listen on the socket.
//
// For unix sockets, the socket will be removed if it already exists.
func (s Socket) Listen() (net.Listener, error) {
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
