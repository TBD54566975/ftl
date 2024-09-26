package identity

import (
	"fmt"
	"strings"
)

type CertificateData struct {
	ID            Identity
	NodePublicKey PublicKey
}

func ParseCertificateData(bytes []byte) (*CertificateData, error) {
	str := string(bytes)
	lastColon := strings.LastIndex(str, ":")
	if lastColon == -1 {
		return nil, fmt.Errorf("no colon in certificate data: %s", str)
	}
	idStr, pubKey := str[:lastColon], str[lastColon+1:]

	id, err := Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse identity: %w", err)
	}

	return &CertificateData{
		ID:            id,
		NodePublicKey: PublicKey(pubKey),
	}, nil
}

func (c CertificateData) String() string {
	return fmt.Sprintf("%s:%x", c.ID, c.NodePublicKey)
}

type Certificate struct {
	SignedData
}

func (c Certificate) String() string {
	return fmt.Sprintf("Certificate %s signed:%x", c.data, c.Signature)
}

// SignCertificateRequest signs an identity certificate request.
// This does not verify the contents of the data--it is assumed that the caller has already done so.
func SignCertificateRequest(caSigner Signer, nodePublic PublicKey, signedData SignedData) (*Certificate, error) {
	nodeVerifier, err := NewTinkVerifier(nodePublic)
	if err != nil {
		return nil, fmt.Errorf("failed to create verifier: %w", err)
	}
	data, err := nodeVerifier.Verify(signedData)
	if err != nil {
		return nil, fmt.Errorf("failed to verify signed data: %w", err)
	}
	id, err := Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse identity: %w", err)
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
	return fmt.Sprintf("CertifiedSignedData data:%s signature:%x (%s)", c.SignedData.data, c.SignedData.Signature, c.Certificate)
}

func (c CertifiedSignedData) Verify(caVerifier Verifier) ([]byte, error) {
	data, err := caVerifier.Verify(c.Certificate.SignedData)
	if err != nil {
		return nil, fmt.Errorf("failed to verify certificate: %w", err)
	}

	certificateData, err := ParseCertificateData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate data: %w", err)
	}

	nodeVerifier, err := NewTinkVerifier(certificateData.NodePublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create verifier: %w", err)
	}
	payload, err := nodeVerifier.Verify(c.SignedData)
	if err != nil {
		return nil, fmt.Errorf("failed to verify signed data: %w", err)
	}

	return payload, nil
}
