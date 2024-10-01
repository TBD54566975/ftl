package identity

import (
	"fmt"

	"google.golang.org/protobuf/proto"

	ftlv1 "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1"
)

type CertificateData struct {
	ID            Identity
	NodePublicKey PublicKey
}

func (c CertificateData) String() string {
	return fmt.Sprintf("CertData(%s %x)", c.ID, c.NodePublicKey)
}

type Certificate struct {
	SignedData
}

func NewCertificate(cert *ftlv1.Certificate) (Certificate, error) {
	data, err := proto.Marshal(cert.Content)
	if err != nil {
		return Certificate{}, fmt.Errorf("failed to marshal certificate: %w", err)
	}

	return Certificate{SignedData: SignedData{data: data, Signature: cert.ControllerSignature}}, nil
}

func (c Certificate) String() string {
	return fmt.Sprintf("Certificate(%x %x)", c.data, c.Signature)
}

func (c Certificate) Proto() *ftlv1.Certificate {
	return &ftlv1.Certificate{
		Content:             &ftlv1.CertificateContent{Identity: c.ID.String()},
		ControllerSignature: c.Signature,
	}
}

// CertifiedSignedData is sent by a node and proves identity based on a certificate.
type CertifiedSignedData struct {
	Certificate Certificate
	SignedData  SignedData
}

func (c CertifiedSignedData) String() string {
	return fmt.Sprintf("CertifiedSignedData data:%x signature:%x (%s)", c.SignedData.data, c.SignedData.Signature, c.Certificate)
}

// Verify against the CA and then the node certificate. Only return the data if both are valid.
func (c CertifiedSignedData) Verify(caVerifier Verifier) (Identity, []byte, error) {
	// Verify against the CA certificate.
	data, err := caVerifier.Verify(c.Certificate.SignedData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to verify certificate: %w", err)
	}

	var certificateContent ftlv1.CertificateContent
	if err = proto.Unmarshal(data, &certificateContent); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal certificate content: %w", err)
	}

	identity, err := Parse(certificateContent.Identity)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse identity: %w", err)
	}

	nodePublicKey := PublicKey{Bytes: certificateContent.PublicKey}
	nodeVerifier, err := NewVerifier(nodePublicKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create verifier: %w", err)
	}
	payload, err := nodeVerifier.Verify(c.SignedData)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to verify signed data: %w", err)
	}

	return identity, payload, nil
}
