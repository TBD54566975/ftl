package identity

import (
	"fmt"

	identitypb "github.com/TBD54566975/ftl/backend/protos/xyz/block/ftl/v1/identity"
)

// CertificateContent is used as a certificate request and also the content of a certificate.
// It is "content" that is to be signed outside of this struct.
type CertificateContent struct {
	Identity  Identity
	PublicKey RawPublicKey
}

func CertificateContentFromProto(proto *identitypb.CertificateContent) (CertificateContent, error) {
	identity, err := Parse(proto.Identity)
	if err != nil {
		return CertificateContent{}, fmt.Errorf("failed to parse identity: %w", err)
	}

	return CertificateContent{
		Identity:  identity,
		PublicKey: NewRawPublicKey(proto.PublicKey),
	}, nil
}

func (c CertificateContent) ToProto() *identitypb.CertificateContent {
	return &identitypb.CertificateContent{
		Identity:  c.Identity.String(),
		PublicKey: c.PublicKey.Bytes,
	}
}

func (c CertificateContent) String() string {
	return fmt.Sprintf("CertficateContent(%s %x)", c.Identity, c.PublicKey)
}

type Certificate struct {
	CertificateContent
	Signature Signature
}

func (c Certificate) Verify(caVerifier Verifier) error {
	signedMessage, err := c.ToSignedMessage()
	if err != nil {
		return fmt.Errorf("failed to encode to signed message: %w", err)
	}
	_, err = caVerifier.Verify(signedMessage)
	if err != nil {
		return fmt.Errorf("failed to verify ca certificate: %w", err)
	}
	return nil
}

func (c Certificate) ToSignedMessage() (SignedMessage, error) {
	// encoded, err := proto.Marshal(&c.CertificateContent.ToProto())
	// if err != nil {
	// 	return SignedMessage{}, fmt.Errorf("failed to marshal certificate content: %w", err)
	// }

	// return SignedMessage{
	// 	message:   encoded,
	// 	Signature: c.Signature,
	// }, nil

	panic("not implemented")
}

func ParseCertificateFromProto(protoCert *identitypb.Certificate) (Certificate, error) {
	// encoded := ParseSignedMessageFromProto(protoCert.SignedMessage)
	// var certificateContent CertificateContent
	// if err := proto.Unmarshal(encoded.message, &certificateContent); err != nil {
	// 	return Certificate{}, fmt.Errorf("failed to unmarshal certificate content: %w", err)
	// }

	// return Certificate{
	// 	CertificateContent: CertificateContent{
	// 		Identity:  certificateContent.Identity,
	// 		PublicKey: certificateContent.PublicKey,
	// 	},
	// 	Signature: encoded.Signature,
	// }, nil
	//
	panic("not implemented")
}

func (c Certificate) String() string {
	return fmt.Sprintf("Certificate(key:%x sig:%x)", c.PublicKey.Bytes, c.Signature.Bytes)
}

func (c Certificate) ToProto() *identitypb.Certificate {
	// certificateContent, err := proto.Marshal(&c.CertificateContent)
	// if err != nil {
	// 	panic(fmt.Errorf("failed to marshal certificate content: %w", err))
	// }

	// return &ftlv1.Certificate{
	// 	SignedMessage: &ftlv1.SignedMessage{
	// 		Message:   certificateContent,
	// 		Signature: c.Signature.Bytes,
	// 	},
	// }

	panic("not implemented")
}

// CertifiedSignedData is sent by a node and proves identity based on a certificate.
type CertifiedSignedData struct {
	Certificate   Certificate
	SignedMessage SignedMessage
}

func (c CertifiedSignedData) String() string {
	return fmt.Sprintf("CertifiedSignedData data:%x signature:%x (%s)", c.SignedMessage.message, c.SignedMessage.Signature, c.Certificate)
}

// Verify against the CA and then the node certificate. Only return the data if both are valid.
func (c CertifiedSignedData) Verify(caVerifier Verifier) (Identity, []byte, error) {
	if err := c.Certificate.Verify(caVerifier); err != nil {
		return nil, nil, fmt.Errorf("failed to verify ca certificate cert:%s: %w", c.Certificate, err)
	}

	// nodePublicKey := RawPublicKey{Bytes: certificateContent.PublicKey}
	// nodeVerifier, err := NewVerifier(nodePublicKey)
	// if err != nil {
	// 	return nil, nil, fmt.Errorf("failed to create verifier: %w", err)
	// }
	// payload, err := nodeVerifier.Verify(c.SignedMessage)
	// if err != nil {
	// 	return nil, nil, fmt.Errorf("failed to verify signed data: %w", err)
	// }

	// return identity, payload, nil
	panic("unimplemented")
}
