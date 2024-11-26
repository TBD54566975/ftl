package identity

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/alecthomas/types/optional"
	"google.golang.org/protobuf/proto"

	v1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
	"github.com/TBD54566975/ftl/internal/model"
)

type Identity interface {
	String() string
}

var _ Identity = Controller{}

type Controller struct{}

func NewController() Controller {
	return Controller{}
}

func (c Controller) String() string {
	return "ca"
}

var _ Identity = Runner{}

// Runner identity
// TODO: Maybe use KeyType[T any, TP keyPayloadConstraint[T]]?
type Runner struct {
	Key        model.RunnerKey
	Deployment string
}

func NewRunner(key model.RunnerKey, module string) Runner {
	return Runner{
		Key:        key,
		Deployment: module,
	}
}

func (r Runner) String() string {
	return fmt.Sprintf("%s:%s", r.Key, r.Deployment)
}

func Parse(s string) (Identity, error) {
	if s == "" {
		return nil, fmt.Errorf("empty identity")
	}
	parts := strings.Split(s, ":")

	if parts[0] == "ca" && len(parts) == 1 {
		return Controller{}, nil
	}

	key, err := model.ParseRunnerKey(parts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse key: %w", err)
	}

	return NewRunner(key, parts[1]), nil
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

	req := &v1.CertificateContent{
		Identity:  s.Identity.String(),
		PublicKey: publicKey.Bytes,
	}
	data, err := proto.Marshal(req)
	if err != nil {
		return v1.GetCertificationRequest{}, fmt.Errorf("failed to marshal cert request: %w", err)
	}

	signed, err := s.Signer.Sign(data)
	if err != nil {
		return v1.GetCertificationRequest{}, fmt.Errorf("failed to sign cert request: %w", err)
	}

	return v1.GetCertificationRequest{
		Request:   req,
		Signature: signed.Signature,
	}, nil
}

func (s *Store) SignCertificateRequest(req *v1.GetCertificationRequest) (Certificate, error) {
	// Ensure the pubkey matches the signature.
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
	_, err = verifier.Verify(signedData)
	if err != nil {
		return Certificate{}, fmt.Errorf("failed to verify signature: %w", err)
	}

	// Discard the node signature as we have verified it.
	// This contains the node's identity and public key.
	certificateData, err := proto.Marshal(req.Request)
	if err != nil {
		return Certificate{}, fmt.Errorf("failed to marshal certificate data: %w", err)
	}
	signedData, err = s.Signer.Sign(certificateData)
	if err != nil {
		return Certificate{}, fmt.Errorf("failed to create ca signed data for cert: %w", err)
	}

	return Certificate{
		SignedData: signedData,
	}, nil
}

func (s *Store) SetCertificate(cert Certificate, controllerVerifier Verifier) error {
	data, err := controllerVerifier.Verify(cert.SignedData)
	if err != nil {
		return fmt.Errorf("failed to verify controller certificate: %w", err)
	}

	// Verify the certificate is for us, checking identity and public key.
	req := &v1.CertificateContent{}
	if err := proto.Unmarshal(data, req); err != nil {
		return fmt.Errorf("failed to unmarshal cert request: %w", err)
	}
	if req.Identity != s.Identity.String() {
		return fmt.Errorf("certificate identity does not match: %s != %s", req.Identity, s.Identity.String())
	}
	myPub, err := s.KeyPair.Public()
	if err != nil {
		return fmt.Errorf("failed to get public key: %w", err)
	}
	if !bytes.Equal(myPub.Bytes, req.PublicKey) {
		return fmt.Errorf("certificate public key does not match")
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
