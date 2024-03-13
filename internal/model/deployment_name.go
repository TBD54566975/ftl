package model

import (
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"fmt"
	"strings"
)

type DeploymentName struct {
	module string
	hash   string
}

var _ interface {
	sql.Scanner
	driver.Valuer
	encoding.TextUnmarshaler
	encoding.TextMarshaler
} = (*DeploymentName)(nil)

func NewDeploymentName(module string) DeploymentName {
	hash := make([]byte, 5)
	_, err := rand.Read(hash)
	if err != nil {
		panic(err)
	}
	return DeploymentName{module: module, hash: fmt.Sprintf("%010x", hash)}
}

func ParseDeploymentName(name string) (DeploymentName, error) {
	if name == "" {
		return DeploymentName{}, nil
	}

	parts := strings.Split(name, "-")
	if len(parts) < 2 {
		return DeploymentName{}, fmt.Errorf("invalid deployment name %q: does not follow <deployment>-<hash> pattern", name)
	}

	module := strings.Join(parts[0:len(parts)-1], "-")
	if len(module) == 0 {
		return DeploymentName{}, fmt.Errorf("invalid deployment name %q: module name should not be empty", name)
	}

	hash := parts[len(parts)-1]
	if len(hash) != 10 {
		return DeploymentName{}, fmt.Errorf("invalid deployment name %q: hash should be 10 hex characters long", name)
	}

	return DeploymentName{
		module: module,
		hash:   parts[len(parts)-1],
	}, nil
}

func (d *DeploymentName) String() string {
	if d.module == "" {
		return ""
	}
	return fmt.Sprintf("%s-%s", d.module, d.hash)
}

func (d *DeploymentName) UnmarshalText(bytes []byte) error {
	name, err := ParseDeploymentName(string(bytes))
	if err != nil {
		return err
	}
	*d = name
	return nil
}

func (d *DeploymentName) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *DeploymentName) Scan(value any) error {
	if value == nil {
		return nil
	}
	name, err := ParseDeploymentName(value.(string))
	if err != nil {
		return err
	}
	*d = name
	return nil
}

func (d DeploymentName) Value() (driver.Value, error) {
	return d.String(), nil
}
