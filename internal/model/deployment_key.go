package model

import (
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"fmt"
	"strings"
)

type DeploymentKey struct {
	module string
	hash   string
}

var _ interface {
	sql.Scanner
	driver.Valuer
	encoding.TextUnmarshaler
	encoding.TextMarshaler
} = (*DeploymentKey)(nil)

func NewDeploymentKey(module string) DeploymentKey {
	hash := make([]byte, 5)
	_, err := rand.Read(hash)
	if err != nil {
		panic(err)
	}
	return DeploymentKey{module: module, hash: fmt.Sprintf("%010x", hash)}
}

func ParseDeploymentKey(input string) (DeploymentKey, error) {
	parts := strings.Split(input, "-")
	if len(parts) < 2 {
		return DeploymentKey{}, fmt.Errorf("invalid deployment key %q: does not follow <deployment>-<hash> pattern", input)
	}

	module := strings.Join(parts[0:len(parts)-1], "-")
	if len(module) == 0 {
		return DeploymentKey{}, fmt.Errorf("invalid deployment key %q: module name should not be empty", input)
	}

	hash := parts[len(parts)-1]
	if len(hash) != 10 {
		return DeploymentKey{}, fmt.Errorf("invalid deployment key %q: hash should be 10 hex characters long", input)
	}

	return DeploymentKey{
		module: module,
		hash:   parts[len(parts)-1],
	}, nil
}

func (d DeploymentKey) String() string {
	return fmt.Sprintf("%s-%s", d.module, d.hash)
}

func (d *DeploymentKey) UnmarshalText(bytes []byte) error {
	key, err := ParseDeploymentKey(string(bytes))
	if err != nil {
		return err
	}
	*d = key
	return nil
}

func (d *DeploymentKey) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *DeploymentKey) Scan(value any) error {
	if value == nil {
		return nil
	}
	key, err := ParseDeploymentKey(value.(string))
	if err != nil {
		return err
	}
	*d = key
	return nil
}

func (d DeploymentKey) Value() (driver.Value, error) {
	return d.String(), nil
}
