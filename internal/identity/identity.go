package identity

import (
	"fmt"
	"strings"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/proto"

	v1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/model"
)

type Identity interface {
	Prefix() string
	String() string
}

func Parse(d string) (Identity, error) {
	parts := strings.Split(d, ":")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid identity: empty string")
	}

	switch parts[0] {
	case "r":
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid runner identity: %s", d)
		}

		runnerKey, err := model.ParseRunnerKey(parts[1])
		if err != nil {
			return nil, fmt.Errorf("failed to parse runner key: %w", err)
		}

		return NewRunner(model.RunnerKey(runnerKey), parts[2]), nil
	case "c":
		if len(parts) != 1 {
			return nil, fmt.Errorf("invalid controller identity: %s", d)
		}

		return Controller{}, nil
	}

	return nil, fmt.Errorf("invalid identity: %s", d)
}

var _ Identity = Runner{}

type Controller struct {
}

func (c Controller) String() string {
	return c.Prefix()
}

func (c Controller) Prefix() string {
	return "c"
}

type Runner struct {
	Key        model.RunnerKey
	Deployment string
}

func NewRunner(key model.RunnerKey, deployment string) Runner {
	return Runner{Key: key, Deployment: deployment}
}

func (r Runner) String() string {
	return fmt.Sprintf("r:%s:%s", r.Key, r.Deployment)
}

func (r Runner) Prefix() string {
	return "r"
}

// Store is held by a node and contains the node's identity, key pair, signer, and certificate.
type Store struct {
	Identity           Identity
	KeyPair            KeyPair
	Signer             Signer
	Certificate        optional.Option[Certificate]
	ControllerVerifier optional.Option[Verifier]
}

func NewStoreNewKeys(identity Identity) (*Store, error) {
	pair, err := GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	signer, err := pair.Signer()
	if err != nil {
		return nil, fmt.Errorf("failed to get signer: %w", err)
	}

	return &Store{
		Identity: identity,
		KeyPair:  pair,
		Signer:   signer,
	}, nil
}

func (s *Store) NewGetCertificateRequest() (v1.GetCertificationRequest, error) {
	publicKey, err := s.KeyPair.Public()
	if err != nil {
		return v1.GetCertificationRequest{}, fmt.Errorf("failed to get public key: %w", err)
	}

	req := v1.CertificationRequest{
		Identity:  s.Identity.String(),
		PublicKey: publicKey.Bytes,
	}
	data, err := proto.Marshal(&req)
	if err != nil {
		return v1.GetCertificationRequest{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	signed, err := s.Signer.Sign(data)
	if err != nil {
		return v1.GetCertificationRequest{}, fmt.Errorf("failed to sign request: %w", err)
	}

	return v1.GetCertificationRequest{
		Request:   &req,
		Signature: signed.Signature,
	}, nil
}

func (s *Store) SignCertificateRequest(req *v1.GetCertificationRequest) (Certificate, error) {
	verifier, err := NewVerifier(NewPublicKey(req.Request.PublicKey))
	if err != nil {
		return Certificate{}, fmt.Errorf("failed to create verifier: %w", err)
	}

	data, err := proto.Marshal(req.Request)
	if err != nil {
		return Certificate{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	signedData, err := NewSignedData(verifier, data, req.Signature)
	if err != nil {
		return Certificate{}, fmt.Errorf("failed to create signed data: %w", err)
	}

	return Certificate{
		SignedData: signedData,
	}, nil
}

func (s *Store) SetCertificate(cert Certificate, controllerVerifier Verifier) error {
	_, err := controllerVerifier.Verify(cert.SignedData)
	if err != nil {
		return fmt.Errorf("failed to verify controller certificate: %w", err)
	}

	s.Certificate = optional.Some(cert)
	s.ControllerVerifier = optional.Some(controllerVerifier)
	return nil
}

func (s *Store) CertifiedSign(data []byte) (CertifiedSignedData, error) {
	certificate, ok := s.Certificate.Get()
	if !ok {
		return CertifiedSignedData{}, fmt.Errorf("certificate not set")
	}

	signedData, err := s.Signer.Sign(data)
	if err != nil {
		return CertifiedSignedData{}, fmt.Errorf("failed to sign data: %w", err)
	}

	return CertifiedSignedData{
		Certificate: certificate,
		SignedData:  signedData,
	}, nil
}
