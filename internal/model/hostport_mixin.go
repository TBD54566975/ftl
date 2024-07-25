package model

import (
	"fmt"
	"strings"

	"github.com/alecthomas/types/optional"
)

// HostPortMixin is a mixin for keys that have an (optional) hostname and a port.
type HostPortMixin struct {
	Hostname optional.Option[string]
	Port     string
}

func (h *HostPortMixin) String() string {
	if hostname, ok := h.Hostname.Get(); ok {
		return hostname + "-" + h.Port
	}
	return h.Port
}

func (h *HostPortMixin) Parse(parts []string) error {
	if len(parts) == 0 {
		return fmt.Errorf("expected <hostname>-<port> but got %q", strings.Join(parts, "-"))
	}
	h.Hostname = optional.Zero(strings.Join(parts[:len(parts)-1], "-"))
	h.Port = parts[len(parts)-1]
	return nil
}

func (h *HostPortMixin) RandomBytes() int { return 10 }
