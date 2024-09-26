package identity

import (
	"encoding/json"
	"fmt"

	"github.com/alecthomas/types/optional"

	"github.com/TBD54566975/ftl/internal/model"
)

type Identity interface {
}

func Parse(d []byte) (Identity, error) {
	var identity Runner // TODO: use protos instead of json
	if json.Unmarshal(d, &identity) != nil {
		return nil, fmt.Errorf("failed to unmarshal identity")
	}

	return identity, nil
}

type Runner struct {
	Key model.RunnerKey
	// Module string
}

func NewRunner(key model.RunnerKey) Runner {
	return Runner{Key: key}
}

func (r Runner) String() string {
	return fmt.Sprintf("r:%s", r.Key)
}

func Sign[T Identity](signer Signer, identity T) (SignedData, error) {
	encoded, err := json.Marshal(identity)
	if err != nil {
		return SignedData{}, fmt.Errorf("failed to marshal identity: %w", err)
	}
	signedData, err := signer.Sign(encoded)
	if err != nil {
		return SignedData{}, fmt.Errorf("failed to sign identity: %w", err)
	}
	return signedData, nil
}

type Store struct {
	Identity    Identity
	KeyPair     KeyPair
	Certificate optional.Option[KeyPair]
}

func NewStore(identity Identity) (*Store, error) {
	// gen signing set
	// keep all in memory
	// optional field for certificate (of the runner's public key)
	// { certificate: { trustedid: "" }, verb: {} }
	// r:<id>:<module> runner
	// c:<id> controller
	// p:<id> provisioner

	pair, err := GenerateTinkKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	return &Store{
		KeyPair: pair,
	}, nil
}
