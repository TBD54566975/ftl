package sqltypes

import (
	"database/sql/driver"
	"strings"

	"github.com/alecthomas/errors"
	"golang.org/x/exp/slices"
)

type Languages []string

func (l Languages) Value() (driver.Value, error) {
	slices.Sort(l)
	return ":" + strings.Join(l, ":") + ":", nil
}

func (l *Languages) Scan(src interface{}) error {
	input, ok := src.(string)
	if !ok {
		return errors.Errorf("expected string, got %T", src)
	}
	if len(input) < 2 {
		return errors.Errorf("expected at least 2 characters, got %d", len(input))
	}
	if input[0] != ':' || input[len(input)-1] != ':' {
		return errors.Errorf("expected leading and trailing colons, got %q", input)
	}
	*l = strings.Split(input[1:len(input)-1], ":")
	return nil
}
