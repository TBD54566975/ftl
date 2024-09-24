package identity

import (
	"fmt"
	"strings"
)

type CertificateData struct {
	ID            Identity
	NodePublicKey PublicKey
}

func ParseCertificateData(s []byte) (CertificateData, error) {
	parts := strings.Split(string(s), ":")
	if len(parts) != 2 {
		return CertificateData{}, fmt.Errorf("invalid certificate data: %s", s)
	}

	id, err := Parse(parts[0])
	if err != nil {
		return CertificateData{}, fmt.Errorf("failed to parse identity: %w", err)
	}

	return CertificateData{
		ID:            id,
		NodePublicKey: PublicKey(parts[1]),
	}, nil
}

func (c CertificateData) String() string {
	return fmt.Sprintf("%s:%x", c.ID, c.NodePublicKey)
}

type Certificate struct {
	SignedData
}

func (c Certificate) String() string {
	return fmt.Sprintf("Certificate %s signed:%x", c.Data, c.Signature)
}

// SignCertificateRequest signs an identity certificate request.
// This does not verify the contents of the data--it is assumed that the caller has already done so.
func SignCertificateRequest(caSigner Signer, nodePublic PublicKey, signedData SignedData) (*Certificate, error) {
	id, err := Parse(string(signedData.Data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse identity: %w", err)
	}

	nodeVerifier, err := NewTinkVerifier(nodePublic)
	if err != nil {
		return nil, fmt.Errorf("failed to create verifier: %w", err)
	}
	err = nodeVerifier.Verify(signedData)
	if err != nil {
		return nil, fmt.Errorf("failed to verify signed data: %w", err)
	}

	certData := CertificateData{
		ID:            id,
		NodePublicKey: nodePublic,
	}

	caSigned, err := caSigner.Sign([]byte(certData.String()))
	if err != nil {
		return nil, fmt.Errorf("failed to sign certificate: %w", err)
	}

	return &Certificate{SignedData: *caSigned}, nil
}

type CertifiedSignedData struct {
	Certificate Certificate
	SignedData  SignedData
}

func (c CertifiedSignedData) String() string {
	return fmt.Sprintf("CertifiedSignedData data:%s signature:%x (%s)", c.SignedData.Data, c.SignedData.Signature, c.Certificate)
}

func (c CertifiedSignedData) Verify(caVerifier Verifier) error {
	certificateData, err := ParseCertificateData(c.SignedData.Data)
	if err != nil {
		return fmt.Errorf("failed to parse identity: %w", err)
	}

	err = caVerifier.Verify(c.Certificate.SignedData)
	if err != nil {
		return fmt.Errorf("failed to verify certificate: %w", err)
	}

	nodeVerifier, err := NewTinkVerifier(c.Certificate.SignedData.Data)
	if err != nil {
		return fmt.Errorf("failed to create verifier: %w", err)
	}
	err = nodeVerifier.Verify(c.SignedData)
	if err != nil {
		return fmt.Errorf("failed to verify signed data: %w", err)
	}

	return nil
}
