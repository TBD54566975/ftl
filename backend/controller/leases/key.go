package leases

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"net/url"
	"strings"
)

// Key is a unique identifier for a lease.
//
// It is a / separated list of strings where each element is URL-path-escaped.
//
// Userspace leases are always in the form "/module/<module>/..." (eg.
// "/module/idv/user/bob"). Internal leases are always in the form "/system/..."
// (eg. "/system/runner/deployment-reservation/<deployment").
//
// Keys should always be created using the SystemKey or ModuleKey functions.
type Key []string

// SystemKey creates an internal system key.
func SystemKey(parts ...string) Key {
	return append([]string{"system"}, parts...)
}

// ModuleKey creates a user-space module key.
func ModuleKey(module string, parts ...string) Key {
	return append([]string{"module", module}, parts...)
}

var _ sql.Scanner = (*Key)(nil)
var _ driver.Valuer = (*Key)(nil)

func (l Key) String() string {
	if len(l) == 0 || (l[0] != "system" && l[0] != "module") {
		panic(fmt.Sprintf("invalid lease key: %#v", l))
	}
	var parts []string
	for _, part := range l {
		parts = append(parts, url.PathEscape(part))
	}
	return "/" + strings.Join(parts, "/")
}

func (l *Key) Scan(dest any) error {
	text, ok := dest.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", dest)
	}
	lk, err := ParseLeaseKey(text)
	if err != nil {
		return err
	}
	*l = lk
	return nil
}

func (l *Key) Value() (driver.Value, error) {
	return l.String(), nil
}

func ParseLeaseKey(s string) (Key, error) {
	if !strings.HasPrefix(s, "/system/") && !strings.HasPrefix(s, "/module/") {
		return nil, fmt.Errorf("invalid lease key: %q", s)
	}
	parts := strings.Split(s, "/")
	for i, part := range parts {
		var err error
		parts[i], err = url.PathUnescape(part)
		if err != nil {
			return nil, err
		}
	}
	return Key(parts[1:]), nil
}
