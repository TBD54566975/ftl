package model

import (
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/hex"
	"fmt"
	"strings"
)

type DeploymentName struct {
	module string
	hash   string
}

// type MaybeDeploymentName optional.Option[DeploymentName]

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
	var zero DeploymentName
	parts := strings.Split(name, "-")
	if len(parts) < 2 {
		return zero, fmt.Errorf("should be at least <deployment>-<hash>: invalid deployment name %q", name)
	}
	hash, err := hex.DecodeString(parts[len(parts)-1])
	if err != nil {
		return zero, fmt.Errorf("invalid deployment name %q: %w", name, err)
	}
	if len(hash) != 5 {
		return zero, fmt.Errorf("hash should be 5 bytes: invalid deployment name %q", name)
	}
	return DeploymentName{
		module: strings.Join(parts[0:len(parts)-1], "-"),
		hash:   fmt.Sprintf("%010x", hash),
	}, nil
}

func (d *DeploymentName) String() string {
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
	fmt.Printf("deploymentName.mashalText(): %s\n", d.String())
	return []byte(d.String()), nil
}

func (d *DeploymentName) Scan(value any) error {
	fmt.Printf("deploymentName.Scan()")
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

func (d *DeploymentName) Value() (driver.Value, error) {
	fmt.Printf("deploymentName.value(): %s\n", d.String())
	return d.String(), nil
}

var _ sql.Scanner = (*DeploymentName)(nil)
var _ driver.Valuer = (*DeploymentName)(nil)
