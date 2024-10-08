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
	String() string
	identity()
}

var _ Identity = Controller{}

type Controller struct{}

func NewController() Controller {
	return Controller{}
}

func (c Controller) String() string {
	return "ca"
}
func (Controller) identity() {}

var _ Identity = Runner{}

// Runner identity
// TODO: Maybe use KeyType[T any, TP keyPayloadConstraint[T]]?
type Runner struct {
	Key    model.RunnerKey
	Module string
}

func NewRunner(key model.RunnerKey, module string) Runner {
	return Runner{
		Key:    key,
		Module: module,
	}
}

func (r Runner) String() string {
	return fmt.Sprintf("%s:%s", r.Key, r.Module)
}

func (Runner) identity() {}

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

func NewStore(identity Identity, keyPair KeyPair) (Store, error) {
	signer, err := keyPair.Signer()
	if err != nil {
		return Store{}, fmt.Errorf("failed to get signer: %w", err)
	}

	return Store{
		Identity: identity,
		KeyPair:  keyPair,
		Signer:   signer,
	}, nil
}

func NewStoreNewKeys(identity Identity) (Store, error) {
	pair, err := GenerateKeyPair()
	if err != nil {
		return Store{}, fmt.Errorf("failed to generate key pair: %w", err)
	}

	signer, err := pair.Signer()
	if err != nil {
		return Store{}, fmt.Errorf("failed to get signer: %w", err)
	}

	return Store{
		Identity: identity,
		KeyPair:  pair,
		Signer:   signer,
	}, nil
}

func (s Store) NewCertificateRequest() (CertificateRequest, error) {
	publicKey, err := s.KeyPair.Public()
	if err != nil {
		return CertificateRequest{}, fmt.Errorf("failed to get public key: %w", err)
	}

	certificateContent := CertificateContent{
		Identity:  s.Identity,
		PublicKey: publicKey,
	}
	encoded, err := proto.Marshal(certificateContent.ToProto())
	if err != nil {
		return CertificateRequest{}, fmt.Errorf("failed to marshal cert content: %w", err)
	}

	signature, err := s.Signer.Sign(encoded)
	if err != nil {
		return CertificateRequest{}, fmt.Errorf("failed to sign cert request: %w", err)
	}

	return CertificateRequest{
		CertificateContent: CertificateContent{
			Identity:  s.Identity,
			PublicKey: publicKey,
		},
		Signature: signature,
	}, nil
}

// // NewGetCertificateRequest generates a signed certificate request to be used in the GetCertificate RPC.
// func (s *Store) ToCertificateRequestProto() (v1.GetCertificateRequest, error) {
// 	publicKey, err := s.KeyPair.Public()
// 	if err != nil {
// 		return v1.GetCertificateRequest{}, fmt.Errorf("failed to get public key: %w", err)
// 	}

// 	content := CertificateContent{
// 		Identity:  s.Identity,
// 		PublicKey: publicKey,
// 	}
// 	message, err := proto.Marshal(content.ToProto())
// 	if err != nil {
// 		return v1.GetCertificateRequest{}, fmt.Errorf("failed to marshal cert content: %w", err)
// 	}

// 	signed, err := s.Signer.Sign(message)
// 	if err != nil {
// 		return v1.GetCertificateRequest{}, fmt.Errorf("failed to sign cert request: %w", err)
// 	}

// 	return v1.GetCertificateRequest{
// 		CertificateRequest: &identitypb.CertificateRequest{
// 			Content:   content.ToProto(),
// 			Signature: signed.Bytes,
// 		},
// 	}, nil
// }

// SignCertificateRequest is called by the controller to sign a certificate request,
// while verifiying the node's signature.
// TODO: Make sure we check the actual given identity!
func (s *Store) SignCertificateRequest(req *v1.GetCertificateRequest) (Certificate, error) {
	// 	signedRequest := ParseSignedMessageFromProto(req.CertificateRequest.SignedMessage)
	// 	encodedRequest := signedRequest.UnverifiedMessage()

	// 	var content CertificateContent
	// 	if err := proto.Unmarshal(encodedRequest, &content); err != nil {
	// 		return Certificate{}, fmt.Errorf("failed to unmarshal certificate content: %w", err)
	// 	}

	// 	// Ensure the given pubkey matches the signature.
	// 	verifier, err := NewVerifier(content.PublicKey)
	// 	if err != nil {
	// 		return Certificate{}, fmt.Errorf("failed to create verifier for pubkey:%x %w", content.PublicKey.Bytes, err)
	// 	}
	// 	_, err = signedRequest.VerifiedMessage(verifier)
	// 	if err != nil {
	// 		return Certificate{}, fmt.Errorf("failed to verify signature: %w", err)
	// 	}

	// 	// Request is valid, sign it.
	// 	signedCertificate, err := s.Signer.Sign(encodedRequest)
	// 	if err != nil {
	// 		return Certificate{}, fmt.Errorf("failed to create ca signed data for cert: %w", err)
	// 	}

	// 	return Certificate{
	// 		CertificateContent: content,
	// 		Signature:          signedCertificate.Signature,
	// 	}, nil
	panic("not implemented")
}

func (s *Store) SetCertificate(cert Certificate, controllerVerifier Verifier) error {
	// data, err := controllerVerifier.Verify(cert.SignedData)
	// if err != nil {
	// 	return fmt.Errorf("failed to verify controller certificate: %w", err)
	// }

	// // Verify the certificate is for us, checking identity and public key.
	// req := &v1.CertificateContent{}
	// if err := proto.Unmarshal(data, req); err != nil {
	// 	return fmt.Errorf("failed to unmarshal cert request: %w", err)
	// }
	// if req.Identity != s.Identity.String() {
	// 	return fmt.Errorf("certificate identity does not match: %s != %s", req.Identity, s.Identity.String())
	// }
	// myPub, err := s.KeyPair.Public()
	// if err != nil {
	// 	return fmt.Errorf("failed to get public key: %w", err)
	// }
	// if !bytes.Equal(myPub.Bytes, req.PublicKey) {
	// 	return fmt.Errorf("certificate public key does not match")
	// }

	// s.Certificate = optional.Some(cert)
	// s.ControllerVerifier = optional.Some(controllerVerifier)
	// return nil
	panic("not implemented")
}

func (s *Store) CertifiedSign(message []byte) (CertifiedSignedData, error) {
	certificate, ok := s.Certificate.Get()
	if !ok {
		return CertifiedSignedData{}, fmt.Errorf("certificate not set")
	}

	signature, err := s.Signer.Sign(message)
	if err != nil {
		return CertifiedSignedData{}, fmt.Errorf("failed to sign data: %w", err)
	}

	return CertifiedSignedData{
		Certificate: certificate,
		Message:     message,
		Signature:   signature,
	}, nil
}
