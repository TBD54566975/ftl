package model

import (
	"crypto/rand"
	"database/sql"
	"database/sql/driver"
	"encoding"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/alecthomas/errors"
	"github.com/alecthomas/types"
)

type DeploymentName string

type MaybeDeploymentName types.Option[DeploymentName]

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
	return DeploymentName(fmt.Sprintf("%s-%s", module, hex.EncodeToString(hash)))
}

func ParseDeploymentName(name string) (DeploymentName, error) {
	var zero DeploymentName
	parts := strings.Split(name, "-")
	if len(parts) < 2 {
		return zero, errors.Errorf("invalid deployment name %q", name)
	}
	hash, err := hex.DecodeString(parts[len(parts)-1])
	if err != nil {
		return zero, errors.Wrapf(err, "invalid deployment name %q", name)
	}
	if len(hash) != 5 {
		return zero, errors.Errorf("invalid deployment name %q", name)
	}
	return DeploymentName(fmt.Sprintf("%s-%s", strings.Join(parts[0:len(parts)-1], "-"), hex.EncodeToString(hash))), nil
}

func (d *DeploymentName) String() string {
	return string(*d)
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
	return []byte(*d), nil
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

func (d *DeploymentName) Value() (driver.Value, error) {
	return d.String(), nil
}
