package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding"
	"errors"
	"strings"
)

type DeploymentKey = KeyType[DeploymentPayload, *DeploymentPayload]

var _ interface {
	sql.Scanner
	driver.Valuer
	encoding.TextUnmarshaler
	encoding.TextMarshaler
} = (*DeploymentKey)(nil)

func NewDeploymentKey(module string) DeploymentKey { return newKey[DeploymentPayload](module) }
func ParseDeploymentKey(key string) (DeploymentKey, error) {
	return parseKey[DeploymentPayload](key)
}

type DeploymentPayload struct {
	Module string
}

var _ KeyPayload = (*DeploymentPayload)(nil)

func (d *DeploymentPayload) Kind() string   { return "dpl" }
func (d *DeploymentPayload) String() string { return d.Module }
func (d *DeploymentPayload) Parse(parts []string) error {
	if len(parts) == 0 {
		return errors.New("expected <module> but got empty string")
	}
	d.Module = strings.Join(parts, "-")
	return nil
}
func (d *DeploymentPayload) RandomBytes() int { return 10 }
