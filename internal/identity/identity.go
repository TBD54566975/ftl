package identity

import (
	"fmt"
	"strings"

	"github.com/alecthomas/types/optional"
)

type Identity interface {
	Name() string
}

func Parse(s string) (Identity, error) {
	parts := strings.Split(s, ":")
	if parts[0] == "r" {
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid runner identity: %s", s)
		}
		return Runner{ID: parts[1], Module: parts[2]}, nil
	}

	return nil, fmt.Errorf("unknown identity: %s", s)
}

var _ Identity = &Runner{}

type Runner struct {
	ID     string
	Module string
}

func (r Runner) Name() string {
	return fmt.Sprintf("r:%s:%s", r.ID, r.Module)
}

func Sign[T Identity](signer Signer, identity T) (*SignedData, error) {
	return signer.Sign([]byte(identity.Name()))
}

type Store struct {
	KeyPair     KeyPair
	Certificate optional.Option[KeyPair]
}

func NewStore() (*Store, error) {
	// gen signing set
	// keep all in memory
	// optional field for certificate (of the runner's public key)
	// { certificate: { trustedid: "" }, verb: {} }
	// runner:<id>:<module>
	// controller:<id>
	// provisioner:<id>

	pair, err := GenerateTinkKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	return &Store{
		KeyPair: pair,
	}, nil
}
